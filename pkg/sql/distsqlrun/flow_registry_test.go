// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package distsqlrun

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/distsqlpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/pkg/errors"
)

// lookupFlow returns the registered flow with the given ID. If no such flow is
// registered, waits until it gets registered - up to the given timeout. If the
// timeout elapses and the flow is not registered, the bool return value will be
// false.
func lookupFlow(fr *flowRegistry, fid distsqlpb.FlowID, timeout time.Duration) *Flow {
	fr.Lock()
	defer fr.Unlock()
	entry := fr.getEntryLocked(fid)
	if entry.flow != nil {
		return entry.flow
	}
	entry = fr.waitForFlowLocked(context.TODO(), fid, timeout)
	if entry == nil {
		return nil
	}
	return entry.flow
}

// lookupStreamInfo returns a stream entry from a flowRegistry. If either the
// flow or the streams are missing, an error is returned.
//
// A copy of the registry's inboundStreamInfo is returned so it can be accessed
// without locking.
func lookupStreamInfo(
	fr *flowRegistry, fid distsqlpb.FlowID, sid distsqlpb.StreamID,
) (inboundStreamInfo, error) {
	fr.Lock()
	defer fr.Unlock()
	entry := fr.getEntryLocked(fid)
	if entry.flow == nil {
		return inboundStreamInfo{}, errors.Errorf("missing flow entry: %s", fid)
	}
	si, ok := entry.inboundStreams[sid]
	if !ok {
		return inboundStreamInfo{}, errors.Errorf("missing stream entry: %d", sid)
	}
	return *si, nil
}

func TestFlowRegistry(t *testing.T) {
	defer leaktest.AfterTest(t)()
	reg := makeFlowRegistry(roachpb.NodeID(0))

	id1 := distsqlpb.FlowID{UUID: uuid.MakeV4()}
	f1 := &Flow{}

	id2 := distsqlpb.FlowID{UUID: uuid.MakeV4()}
	f2 := &Flow{}

	id3 := distsqlpb.FlowID{UUID: uuid.MakeV4()}
	f3 := &Flow{}

	id4 := distsqlpb.FlowID{UUID: uuid.MakeV4()}
	f4 := &Flow{}

	// A basic duration; needs to be significantly larger than possible delays
	// in scheduling goroutines.
	jiffy := 10 * time.Millisecond

	// -- Lookup, register, lookup, unregister, lookup. --

	if f := lookupFlow(reg, id1, 0); f != nil {
		t.Error("looked up unregistered flow")
	}

	const flowStreamTimeout = 10 * time.Second

	ctx := context.Background()
	if err := reg.RegisterFlow(
		ctx, id1, f1, nil /* inboundStreams */, flowStreamTimeout,
	); err != nil {
		t.Fatal(err)
	}

	if f := lookupFlow(reg, id1, 0); f != f1 {
		t.Error("couldn't lookup previously registered flow")
	}

	reg.UnregisterFlow(id1)

	if f := lookupFlow(reg, id1, 0); f != nil {
		t.Error("looked up unregistered flow")
	}

	// -- Lookup with timeout, register in the meantime. --

	go func() {
		time.Sleep(jiffy)
		if err := reg.RegisterFlow(
			ctx, id1, f1, nil /* inboundStreams */, flowStreamTimeout,
		); err != nil {
			t.Error(err)
		}
	}()

	if f := lookupFlow(reg, id1, 10*jiffy); f != f1 {
		t.Error("couldn't lookup registered flow (with wait)")
	}

	if f := lookupFlow(reg, id1, 0); f != f1 {
		t.Error("couldn't lookup registered flow")
	}

	// -- Multiple lookups before register. --

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if f := lookupFlow(reg, id2, 10*jiffy); f != f2 {
			t.Error("couldn't lookup registered flow (with wait)")
		}
		wg.Done()
	}()

	go func() {
		if f := lookupFlow(reg, id2, 10*jiffy); f != f2 {
			t.Error("couldn't lookup registered flow (with wait)")
		}
		wg.Done()
	}()

	time.Sleep(jiffy)
	if err := reg.RegisterFlow(
		ctx, id2, f2, nil /* inboundStreams */, flowStreamTimeout,
	); err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// -- Multiple lookups, with the first one failing. --

	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	wg1.Add(1)
	wg2.Add(1)
	go func() {
		if f := lookupFlow(reg, id3, jiffy); f != nil {
			t.Error("expected lookup to fail")
		}
		wg1.Done()
	}()

	go func() {
		if f := lookupFlow(reg, id3, 10*jiffy); f != f3 {
			t.Error("couldn't lookup registered flow (with wait)")
		}
		wg2.Done()
	}()

	wg1.Wait()
	if err := reg.RegisterFlow(
		ctx, id3, f3, nil /* inboundStreams */, flowStreamTimeout,
	); err != nil {
		t.Fatal(err)
	}
	wg2.Wait()

	// -- Lookup with huge timeout, register in the meantime. --

	go func() {
		time.Sleep(jiffy)
		if err := reg.RegisterFlow(
			ctx, id4, f4, nil /* inboundStreams */, flowStreamTimeout,
		); err != nil {
			t.Error(err)
		}
	}()

	// This should return in a jiffy.
	if f := lookupFlow(reg, id4, time.Hour); f != f4 {
		t.Error("couldn't lookup registered flow (with wait)")
	}
}

// Test that, if inbound streams are not connected within the timeout, errors
// are propagated to their consumers and future attempts to connect them fail.
func TestStreamConnectionTimeout(t *testing.T) {
	defer leaktest.AfterTest(t)()
	reg := makeFlowRegistry(roachpb.NodeID(0))

	jiffy := time.Nanosecond

	// Register a flow with a very low timeout. After it times out, we'll attempt
	// to connect a stream, but it'll be too late.
	id1 := distsqlpb.FlowID{UUID: uuid.MakeV4()}
	f1 := &Flow{}
	streamID1 := distsqlpb.StreamID(1)
	consumer := &RowBuffer{}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	inboundStreams := map[distsqlpb.StreamID]*inboundStreamInfo{
		streamID1: {receiver: consumer, waitGroup: wg},
	}
	if err := reg.RegisterFlow(
		context.TODO(), id1, f1, inboundStreams, jiffy,
	); err != nil {
		t.Fatal(err)
	}

	testutils.SucceedsSoon(t, func() error {
		si, err := lookupStreamInfo(reg, id1, streamID1)
		if err != nil {
			t.Fatal(err)
		}
		if !si.canceled {
			return errors.Errorf("not timed out yet")
		}
		return nil
	})

	testutils.SucceedsSoon(t, func() error {
		if !consumer.ProducerClosed() {
			return errors.New("expected consumer to have been closed when the flow timed out")
		}
		return nil
	})

	// Create a dummy server stream to pass to ConnectInboundStream.
	serverStream, _ /* clientStream */, cleanup, err := createDummyStream()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	_, _, _, err = reg.ConnectInboundStream(context.TODO(), id1, streamID1, serverStream, jiffy)
	if !testutils.IsError(err, "came too late") {
		t.Fatalf("expected %q, got: %v", "came too late", err)
	}

	// Unregister the flow. Subsequent attempts to connect a stream should result
	// in a different error than before.
	reg.UnregisterFlow(id1)
	_, _, _, err = reg.ConnectInboundStream(context.TODO(), id1, streamID1, serverStream, jiffy)
	if !testutils.IsError(err, "not found") {
		t.Fatalf("expected %q, got: %v", "not found", err)
	}
}

// Test that the FlowRegistry send the correct handshake messages:
// - if an inbound stream arrives to the registry before the consumer is
// scheduled, then a Handshake message informing that the consumer is not yet
// connected is sent;
// - once the consumer connects, another Handshake message is sent.
func TestHandshake(t *testing.T) {
	defer leaktest.AfterTest(t)()

	reg := makeFlowRegistry(roachpb.NodeID(0))

	tests := []struct {
		name                   string
		consumerConnectedEarly bool
	}{
		{
			name:                   "consumer early",
			consumerConnectedEarly: true,
		},
		{
			name:                   "consumer late",
			consumerConnectedEarly: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flowID := distsqlpb.FlowID{UUID: uuid.MakeV4()}
			streamID := distsqlpb.StreamID(1)

			serverStream, clientStream, cleanup, err := createDummyStream()
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			connectProducer := func() {
				// Simulate a producer connecting to the server. This should be called
				// async because the consumer is not yet there and ConnectInboundStream
				// is blocking.
				if _, _, _, err := reg.ConnectInboundStream(
					context.TODO(), flowID, streamID, serverStream, time.Hour,
				); err != nil {
					t.Error(err)
				}
			}
			connectConsumer := func() {
				f1 := &Flow{}
				consumer := &RowBuffer{}
				wg := &sync.WaitGroup{}
				wg.Add(1)
				inboundStreams := map[distsqlpb.StreamID]*inboundStreamInfo{
					streamID: {receiver: consumer, waitGroup: wg},
				}
				if err := reg.RegisterFlow(
					context.TODO(), flowID, f1, inboundStreams, time.Hour, /* timeout */
				); err != nil {
					t.Fatal(err)
				}
			}

			// If the consumer is supposed to be connected early, then we connect the
			// consumer and then we connect the producer. Otherwise, we connect the
			// producer and expect a first handshake and only then we connect the
			// consumer.
			if tc.consumerConnectedEarly {
				connectConsumer()
				go connectProducer()
			} else {
				go connectProducer()

				// Expect the client (the producer) to receive a Handshake saying that the
				// consumer is not connected yet.
				consumerSignal, err := clientStream.Recv()
				if err != nil {
					t.Fatal(err)
				}
				if consumerSignal.Handshake == nil {
					t.Fatalf("expected handshake, got: %+v", consumerSignal)
				}
				if consumerSignal.Handshake.ConsumerScheduled {
					t.Fatal("expected !ConsumerScheduled")
				}

				connectConsumer()
			}

			// Now expect another Handshake message telling the producer that the consumer
			// has connected.
			consumerSignal, err := clientStream.Recv()
			if err != nil {
				t.Fatal(err)
			}
			if consumerSignal.Handshake == nil {
				t.Fatalf("expected handshake, got: %+v", consumerSignal)
			}
			if !consumerSignal.Handshake.ConsumerScheduled {
				t.Fatal("expected ConsumerScheduled")
			}
		})
	}
}

// TestFlowRegistryDrain verifies a flowRegistry's draining behavior. See
// subtests for more details.
func TestFlowRegistryDrain(t *testing.T) {
	defer leaktest.AfterTest(t)()

	ctx := context.TODO()
	reg := makeFlowRegistry(roachpb.NodeID(0))

	flow := &Flow{}
	id := distsqlpb.FlowID{UUID: uuid.MakeV4()}
	registerFlow := func(t *testing.T, id distsqlpb.FlowID) {
		t.Helper()
		if err := reg.RegisterFlow(
			ctx, id, flow, nil /* inboundStreams */, 0, /* timeout */
		); err != nil {
			t.Fatal(err)
		}
	}

	// WaitForFlow verifies that Drain waits for a flow to finish within the
	// timeout.
	t.Run("WaitForFlow", func(t *testing.T) {
		registerFlow(t, id)
		drainDone := make(chan struct{})
		go func() {
			reg.Drain(math.MaxInt64 /* flowDrainWait */, 0 /* minFlowDrainWait */)
			drainDone <- struct{}{}
		}()
		// Be relatively sure that the flowRegistry is draining.
		time.Sleep(time.Microsecond)
		reg.UnregisterFlow(id)
		<-drainDone
		reg.Undrain()
	})

	// DrainTimeout verifies that Drain returns once the timeout expires.
	t.Run("DrainTimeout", func(t *testing.T) {
		registerFlow(t, id)
		reg.Drain(0 /* flowDrainWait */, 0 /* minFlowDrainWait */)
		reg.UnregisterFlow(id)
		reg.Undrain()
	})

	// AcceptNewFlow verifies that a flowRegistry continues accepting flows
	// while draining.
	t.Run("AcceptNewFlow", func(t *testing.T) {
		registerFlow(t, id)
		drainDone := make(chan struct{})
		go func() {
			reg.Drain(math.MaxInt64 /* flowDrainWait */, 0 /* minFlowDrainWait */)
			drainDone <- struct{}{}
		}()
		// Be relatively sure that the flowRegistry is draining.
		time.Sleep(time.Microsecond)
		newFlowID := distsqlpb.FlowID{UUID: uuid.MakeV4()}
		registerFlow(t, newFlowID)
		reg.UnregisterFlow(id)
		select {
		case <-drainDone:
			t.Fatal("finished draining before unregistering new flow")
		default:
		}
		reg.UnregisterFlow(newFlowID)
		<-drainDone
		// The registry should not accept new flows once it has finished draining.
		if err := reg.RegisterFlow(
			ctx, id, flow, nil /* inboundStreams */, 0, /* timeout */
		); !testutils.IsError(err, "draining") {
			t.Fatalf("unexpected error: %v", err)
		}
		reg.Undrain()
	})

	// MinFlowWait verifies that the flowRegistry waits a minimum amount of time
	// for incoming flows to be registered.
	t.Run("MinFlowWait", func(t *testing.T) {
		// Case in which draining is initiated with zero running flows.
		drainDone := make(chan struct{})
		// Register a flow right before the flowRegistry waits for
		// minFlowDrainWait. Use an errChan because draining is performed from
		// another goroutine and cannot call t.Fatal.
		errChan := make(chan error)
		reg.testingRunBeforeDrainSleep = func() {
			if err := reg.RegisterFlow(
				ctx, id, flow, nil /* inboundStreams */, 0, /* timeout */
			); err != nil {
				errChan <- err
			}
			errChan <- nil
		}
		defer func() { reg.testingRunBeforeDrainSleep = nil }()
		go func() {
			reg.Drain(math.MaxInt64 /* flowDrainWait */, 0 /* minFlowDrainWait */)
			drainDone <- struct{}{}
		}()
		if err := <-errChan; err != nil {
			t.Fatal(err)
		}
		reg.UnregisterFlow(id)
		<-drainDone
		reg.Undrain()

		// Case in which a running flow finishes before the minimum wait time. We
		// attempt to register another flow after the completion of the first flow
		// to simulate an incoming flow that is registered during the minimum wait
		// time. However, it is possible to unregister the first flow after the
		// minimum wait time has passed, in which case we simply verify that the
		// flowRegistry drain process has lasted at least the required wait time.
		registerFlow(t, id)
		reg.testingRunBeforeDrainSleep = func() {
			if err := reg.RegisterFlow(
				ctx, id, flow, nil /* inboundStreams */, 0, /* timeout */
			); err != nil {
				errChan <- err
			}
			errChan <- nil
		}
		minFlowDrainWait := 10 * time.Millisecond
		start := timeutil.Now()
		go func() {
			reg.Drain(math.MaxInt64 /* flowDrainWait */, minFlowDrainWait)
			drainDone <- struct{}{}
		}()
		// Be relatively sure that the flowRegistry is draining.
		time.Sleep(time.Microsecond)
		reg.UnregisterFlow(id)
		select {
		case <-drainDone:
			if timeutil.Since(start) < minFlowDrainWait {
				t.Fatal("flow registry did not wait at least minFlowDrainWait")
			}
			return
		case err := <-errChan:
			if err != nil {
				t.Fatal(err)
			}
		}
		reg.UnregisterFlow(id)
		<-drainDone

		reg.Undrain()
	})
}

// Test that we can register send a sync flow to the distSQLSrv after the
// flowRegistry is draining and the we can also clean that flow up (the flow
// will get a draining error). This used to crash.
func TestSyncFlowAfterDrain(t *testing.T) {
	defer leaktest.AfterTest(t)()

	ctx := context.TODO()
	// We'll create a server just so that we can extract its distsql ServerConfig,
	// so we can use it for a manually-built DistSQL Server below. Otherwise, too
	// much work to create that ServerConfig by hand.
	s, _, _ := serverutils.StartServer(t, base.TestServerArgs{})
	defer s.Stopper().Stop(ctx)
	cfg := s.DistSQLServer().(*ServerImpl).ServerConfig

	distSQLSrv := NewServer(ctx, cfg)
	distSQLSrv.flowRegistry.Drain(time.Duration(0) /* flowDrainWait */, time.Duration(0) /* minFlowDrainWait */)

	// We create some flow; it doesn't matter what.
	req := distsqlpb.SetupFlowRequest{Version: Version}
	req.Flow = distsqlpb.FlowSpec{
		Processors: []distsqlpb.ProcessorSpec{
			{
				Core: distsqlpb.ProcessorCoreUnion{Values: &distsqlpb.ValuesCoreSpec{}},
				Output: []distsqlpb.OutputRouterSpec{{
					Type:    distsqlpb.OutputRouterSpec_PASS_THROUGH,
					Streams: []distsqlpb.StreamEndpointSpec{{StreamID: 1, Type: distsqlpb.StreamEndpointSpec_REMOTE}},
				}},
			},
			{
				Input: []distsqlpb.InputSyncSpec{{
					Type:    distsqlpb.InputSyncSpec_UNORDERED,
					Streams: []distsqlpb.StreamEndpointSpec{{StreamID: 1, Type: distsqlpb.StreamEndpointSpec_REMOTE}},
				}},
				Core: distsqlpb.ProcessorCoreUnion{Noop: &distsqlpb.NoopCoreSpec{}},
				Output: []distsqlpb.OutputRouterSpec{{
					Type:    distsqlpb.OutputRouterSpec_PASS_THROUGH,
					Streams: []distsqlpb.StreamEndpointSpec{{Type: distsqlpb.StreamEndpointSpec_SYNC_RESPONSE}},
				}},
			},
		},
	}

	types := make([]types.T, 0)
	rb := NewRowBuffer(types, nil /* rows */, RowBufferArgs{})
	ctx, flow, err := distSQLSrv.SetupSyncFlow(ctx, &distSQLSrv.memMonitor, &req, rb)
	if err != nil {
		t.Fatal(err)
	}
	if err := flow.Start(ctx, func() {}); err != nil {
		t.Fatal(err)
	}
	flow.Wait()
	_, meta := rb.Next()
	if meta == nil {
		t.Fatal("expected draining err, got no meta")
	}
	if !testutils.IsError(meta.Err, "the registry is draining") {
		t.Fatalf("expected draining err, got: %v", meta.Err)
	}
	flow.Cleanup(ctx)
}

// TestInboundStreamTimeoutIsRetryable verifies that a failure from an inbound
// stream to connect in a timeout is considered retryable by
// pgerror.IsSQLRetryableError.
// TODO(asubiotto): This error should also be considered retryable by clients.
func TestInboundStreamTimeoutIsRetryable(t *testing.T) {
	defer leaktest.AfterTest(t)()

	fr := makeFlowRegistry(0)
	wg := sync.WaitGroup{}
	rc := &RowChannel{}
	rc.initWithBufSizeAndNumSenders(sqlbase.OneIntCol, 1 /* chanBufSize */, 1 /* numSenders */)
	inboundStreams := map[distsqlpb.StreamID]*inboundStreamInfo{
		0: {
			receiver:  rc,
			waitGroup: &wg,
		},
	}
	wg.Add(1)
	if err := fr.RegisterFlow(
		context.Background(), distsqlpb.FlowID{}, &Flow{}, inboundStreams, 0, /* timeout */
	); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if _, meta := rc.Next(); meta == nil {
		t.Fatal("expected error but got no meta")
	} else if !pgerror.IsSQLRetryableError(meta.Err) {
		t.Fatalf("unexpected error: %v", meta.Err)
	}
}

// TestTimeoutPushDoesntBlockRegister verifies that in the case of a timeout
// error, we are still able to register flows while Pushing the error (#34041).
func TestTimeoutPushDoesntBlockRegister(t *testing.T) {
	defer leaktest.AfterTest(t)()

	ctx := context.Background()
	fr := makeFlowRegistry(0)
	// pushChan is used to be able to tell when a Push on the RowBuffer has
	// occurred.
	pushChan := make(chan *distsqlpb.ProducerMetadata)
	rc := NewRowBuffer(
		sqlbase.OneIntCol,
		nil, /* rows */
		RowBufferArgs{
			OnPush: func(_ sqlbase.EncDatumRow, meta *distsqlpb.ProducerMetadata) {
				pushChan <- meta
				<-pushChan
			},
		},
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	inboundStreams := map[distsqlpb.StreamID]*inboundStreamInfo{
		0: {
			receiver:  rc,
			waitGroup: &wg,
		},
	}

	// RegisterFlow with an immediate timeout.
	if err := fr.RegisterFlow(
		ctx, distsqlpb.FlowID{}, &Flow{}, inboundStreams, 0, /* timeout */
	); err != nil {
		t.Fatal(err)
	}

	// Ensure RegisterFlow performs a Push.
	meta := <-pushChan
	if !testutils.IsError(meta.Err, errNoInboundStreamConnection.Error()) {
		t.Fatalf("unexpected err %v, expected %s", meta.Err, errNoInboundStreamConnection)
	}

	// Attempt to register a flow. Note that this flow has no inbound streams, so
	// Pushing to the RowBuffer is unexpected.
	if err := fr.RegisterFlow(
		ctx, distsqlpb.FlowID{UUID: uuid.MakeV4()}, &Flow{}, nil /* inboundStreams */, time.Hour, /* timeout */
	); err != nil {
		t.Fatal(err)
	}

	// Unblock the first RegisterFlow.
	close(pushChan)
}

// TestFlowCancelPartiallyBlocked tests that cancellation messages can propagate
// into a flow even if one of the inbound streams are blocked (#35859).
func TestFlowCancelPartiallyBlocked(t *testing.T) {
	defer leaktest.AfterTest(t)()

	ctx := context.Background()
	fr := makeFlowRegistry(0)
	left := &RowChannel{}
	left.initWithBufSizeAndNumSenders(nil /* types */, 1, 1)
	right := &RowChannel{}
	right.initWithBufSizeAndNumSenders(nil /* types */, 1, 1)

	wgLeft := sync.WaitGroup{}
	wgLeft.Add(1)
	wgRight := sync.WaitGroup{}
	wgRight.Add(1)
	inboundStreams := map[distsqlpb.StreamID]*inboundStreamInfo{
		0: {
			receiver:  left,
			waitGroup: &wgLeft,
		},
		1: {
			receiver:  right,
			waitGroup: &wgRight,
		},
	}

	// Fill up the left, so pushes to it block.
	left.Push(nil, &distsqlpb.ProducerMetadata{})

	// RegisterFlow with an immediate timeout.
	flow := &Flow{
		FlowCtx: FlowCtx{
			id: distsqlpb.FlowID{UUID: uuid.FastMakeV4()},
		},
		inboundStreams: inboundStreams,
		flowRegistry:   fr,
	}
	if err := fr.RegisterFlow(
		ctx, flow.id, flow, inboundStreams, 10*time.Second, /* timeout */
	); err != nil {
		t.Fatal(err)
	}

	flow.cancel()

	// Reading from the right shouldn't block and should immediately return a
	// flow canceled error.

	_, meta := right.Next()
	if meta.Err != sqlbase.QueryCanceledError {
		t.Fatal("expected query canceled, found", meta.Err)
	}

	// Read from the left to unblock the canceler, assert that the next
	// message is the query canceled message as well.

	_, _ = left.Next()
	_, meta = left.Next()
	if meta.Err != sqlbase.QueryCanceledError {
		t.Fatal("expected query canceled, found", meta.Err)
	}
}
