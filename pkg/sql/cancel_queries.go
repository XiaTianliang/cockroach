// Copyright 2017 The Cockroach Authors.
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

package sql

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/errors"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type cancelQueriesNode struct {
	rows     planNode
	ifExists bool
}

func (p *planner) CancelQueries(ctx context.Context, n *tree.CancelQueries) (planNode, error) {
	rows, err := p.newPlan(ctx, n.Queries, []*types.T{types.String})
	if err != nil {
		return nil, err
	}
	cols := planColumns(rows)
	if len(cols) != 1 {
		return nil, pgerror.Newf(pgerror.CodeSyntaxError,
			"CANCEL QUERIES expects a single column source, got %d columns", len(cols))
	}
	if !cols[0].Typ.Equivalent(types.String) {
		return nil, pgerror.Newf(pgerror.CodeDatatypeMismatchError,
			"CANCEL QUERIES requires string values, not type %s", cols[0].Typ)
	}

	return &cancelQueriesNode{
		rows:     rows,
		ifExists: n.IfExists,
	}, nil
}

func (n *cancelQueriesNode) startExec(runParams) error {
	return nil
}

func (n *cancelQueriesNode) Next(params runParams) (bool, error) {
	// TODO(knz): instead of performing the cancels sequentially,
	// accumulate all the query IDs and then send batches to each of the
	// nodes.

	if ok, err := n.rows.Next(params); err != nil || !ok {
		return ok, err
	}

	datum := n.rows.Values()[0]
	if datum == tree.DNull {
		return true, nil
	}

	statusServer := params.extendedEvalCtx.StatusServer
	queryIDString, ok := tree.AsDString(datum)
	if !ok {
		return false, pgerror.AssertionFailedf("%q: expected *DString, found %T", datum, datum)
	}

	queryID, err := StringToClusterWideID(string(queryIDString))
	if err != nil {
		return false, pgerror.Wrapf(err, pgerror.CodeSyntaxError, "invalid query ID %s", datum)
	}

	// Get the lowest 32 bits of the query ID.
	nodeID := 0xFFFFFFFF & queryID.Lo

	request := &serverpb.CancelQueryRequest{
		NodeId:   fmt.Sprintf("%d", nodeID),
		QueryID:  string(queryIDString),
		Username: params.SessionData().User,
	}

	response, err := statusServer.CancelQuery(params.ctx, request)
	if err != nil {
		return false, err
	}

	if !response.Canceled && !n.ifExists {
		return false, errors.Newf("could not cancel query %s: %s", queryID, response.Error)
	}

	return true, nil
}

func (*cancelQueriesNode) Values() tree.Datums { return nil }

func (n *cancelQueriesNode) Close(ctx context.Context) {
	n.rows.Close(ctx)
}
