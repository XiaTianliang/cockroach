// Copyright 2018 The Cockroach Authors.
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

	"github.com/cockroachdb/cockroach/pkg/sql/distsqlpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
)

// projectSetProcessor is the physical processor implementation of
// projectSetNode.
type projectSetProcessor struct {
	ProcessorBase

	input RowSource
	spec  *distsqlpb.ProjectSetSpec

	// exprHelpers are the constant-folded, type checked expressions specified
	// in the ROWS FROM syntax. This can contain many kinds of expressions
	// (anything that is "function-like" including COALESCE, NULLIF) not just
	// SRFs.
	exprHelpers []*exprHelper

	// funcs contains a valid pointer to a SRF FuncExpr for every entry
	// in `exprHelpers` that is actually a SRF function application.
	// The size of the slice is the same as `exprHelpers` though.
	funcs []*tree.FuncExpr

	// inputRowReady is set when there was a row of input data available
	// from the source.
	inputRowReady bool

	// rowBuffer will contain the current row of results.
	rowBuffer sqlbase.EncDatumRow

	// gens contains the current "active" ValueGenerators for each entry
	// in `funcs`. They are initialized anew for every new row in the source.
	gens []tree.ValueGenerator

	// done indicates for each `expr` whether the values produced by
	// either the SRF or the scalar expressions are fully consumed and
	// thus also whether NULLs should be emitted instead.
	done []bool

	// emitCount is used to track the number of rows that have been
	// emitted from Next().
	emitCount int64
}

var _ Processor = &projectSetProcessor{}
var _ RowSource = &projectSetProcessor{}

const projectSetProcName = "projectSet"

func newProjectSetProcessor(
	flowCtx *FlowCtx,
	processorID int32,
	spec *distsqlpb.ProjectSetSpec,
	input RowSource,
	post *distsqlpb.PostProcessSpec,
	output RowReceiver,
) (*projectSetProcessor, error) {
	outputTypes := append(input.OutputTypes(), spec.GeneratedColumns...)
	ps := &projectSetProcessor{
		input:       input,
		spec:        spec,
		exprHelpers: make([]*exprHelper, len(spec.Exprs)),
		funcs:       make([]*tree.FuncExpr, len(spec.Exprs)),
		rowBuffer:   make(sqlbase.EncDatumRow, len(outputTypes)),
		gens:        make([]tree.ValueGenerator, len(spec.Exprs)),
		done:        make([]bool, len(spec.Exprs)),
	}
	if err := ps.Init(
		ps,
		post,
		outputTypes,
		flowCtx,
		processorID,
		output,
		nil, /* memMonitor */
		ProcStateOpts{InputsToDrain: []RowSource{ps.input}},
	); err != nil {
		return nil, err
	}
	return ps, nil
}

// Start is part of the RowSource interface.
func (ps *projectSetProcessor) Start(ctx context.Context) context.Context {
	ps.input.Start(ctx)
	ctx = ps.StartInternal(ctx, projectSetProcName)

	// Initialize exprHelpers.
	for i, expr := range ps.spec.Exprs {
		var helper exprHelper
		err := helper.init(expr, ps.input.OutputTypes(), ps.evalCtx)
		if err != nil {
			ps.MoveToDraining(err)
			return ctx
		}
		if tFunc, ok := helper.expr.(*tree.FuncExpr); ok && tFunc.IsGeneratorApplication() {
			// expr is a set-generating function.
			ps.funcs[i] = tFunc
		}
		ps.exprHelpers[i] = &helper
	}
	return ctx
}

// nextInputRow returns the next row or metadata from ps.input. It also
// initializes the value generators for that row.
func (ps *projectSetProcessor) nextInputRow() (
	sqlbase.EncDatumRow,
	*distsqlpb.ProducerMetadata,
	error,
) {
	row, meta := ps.input.Next()
	if row == nil {
		return nil, meta, nil
	}

	// Initialize a round of SRF generators or scalar values.
	for i := range ps.exprHelpers {
		if fn := ps.funcs[i]; fn != nil {
			// A set-generating function. Prepare its ValueGenerator.

			// Set exprHelper.row so that we can use it as an IndexedVarContainer.
			ps.exprHelpers[i].row = row

			ps.evalCtx.IVarContainer = ps.exprHelpers[i]
			gen, err := fn.EvalArgsAndGetGenerator(ps.evalCtx)
			if err != nil {
				return nil, nil, err
			}
			if gen == nil {
				gen = builtins.EmptyGenerator()
			}
			if err := gen.Start(); err != nil {
				return nil, nil, err
			}
			ps.gens[i] = gen
		}
		ps.done[i] = false
	}

	return row, nil, nil
}

// nextGeneratorValues populates the row buffer with the next set of generated
// values. It returns true if any of the generators produce new values.
func (ps *projectSetProcessor) nextGeneratorValues() (newValAvail bool, err error) {
	colIdx := len(ps.input.OutputTypes())
	for i := range ps.exprHelpers {
		// Do we have a SRF?
		if gen := ps.gens[i]; gen != nil {
			// Yes. Is there still work to do for the current row?
			numCols := int(ps.spec.NumColsPerGen[i])
			if !ps.done[i] {
				// Yes; check whether this source still has some values available.
				hasVals, err := gen.Next()
				if err != nil {
					return false, err
				}
				if hasVals {
					// This source has values, use them.
					for _, value := range gen.Values() {
						ps.rowBuffer[colIdx] = ps.toEncDatum(value, colIdx)
						colIdx++
					}
					newValAvail = true
				} else {
					ps.done[i] = true
					// No values left. Fill the buffer with NULLs for future results.
					for j := 0; j < numCols; j++ {
						ps.rowBuffer[colIdx] = ps.toEncDatum(tree.DNull, colIdx)
						colIdx++
					}
				}
			} else {
				// Already done. Increment colIdx.
				colIdx += numCols
			}
		} else {
			// A simple scalar result.
			// Do we still need to produce the scalar value? (first row)
			if !ps.done[i] {
				// Yes. Produce it once, then indicate it's "done".
				value, err := ps.exprHelpers[i].eval(ps.rowBuffer)
				if err != nil {
					return false, err
				}
				ps.rowBuffer[colIdx] = ps.toEncDatum(value, colIdx)
				colIdx++
				newValAvail = true
				ps.done[i] = true
			} else {
				// Ensure that every row after the first returns a NULL value.
				ps.rowBuffer[colIdx] = ps.toEncDatum(tree.DNull, colIdx)
				colIdx++
			}
		}
	}
	return newValAvail, nil
}

// Next is part of the RowSource interface.
func (ps *projectSetProcessor) Next() (sqlbase.EncDatumRow, *distsqlpb.ProducerMetadata) {
	const cancelCheckCount = 10000

	for ps.State == StateRunning {

		// Occasionally check for cancellation.
		ps.emitCount++
		if ps.emitCount%cancelCheckCount == 0 {
			if err := ps.Ctx.Err(); err != nil {
				ps.MoveToDraining(err)
				return nil, ps.DrainHelper()
			}
		}

		// Start of a new row of input?
		if !ps.inputRowReady {
			// Read the row from the source.
			row, meta, err := ps.nextInputRow()
			if meta != nil {
				if meta.Err != nil {
					ps.MoveToDraining(nil /* err */)
				}
				return nil, meta
			}
			if err != nil {
				ps.MoveToDraining(err)
				return nil, ps.DrainHelper()
			}
			if row == nil {
				ps.MoveToDraining(nil /* err */)
				return nil, ps.DrainHelper()
			}

			// Keep the values for later.
			copy(ps.rowBuffer, row)
			ps.inputRowReady = true
		}

		// Try to find some data on the generator side.
		newValAvail, err := ps.nextGeneratorValues()
		if err != nil {
			ps.MoveToDraining(err)
			return nil, ps.DrainHelper()
		}
		if newValAvail {
			if outRow := ps.ProcessRowHelper(ps.rowBuffer); outRow != nil {
				return outRow, nil
			}
		} else {
			// The current batch of SRF values was exhausted. Advance
			// to the next input row.
			ps.inputRowReady = false
		}
	}
	return nil, ps.DrainHelper()
}

func (ps *projectSetProcessor) toEncDatum(d tree.Datum, colIdx int) sqlbase.EncDatum {
	generatedColIdx := colIdx - len(ps.input.OutputTypes())
	ctyp := &ps.spec.GeneratedColumns[generatedColIdx]
	return sqlbase.DatumToEncDatum(ctyp, d)
}

// ConsumerClosed is part of the RowSource interface.
func (ps *projectSetProcessor) ConsumerClosed() {
	// The consumer is done, Next() will not be called again.
	ps.InternalClose()
}
