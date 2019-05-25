// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in licenses/BSD-vitess.txt.

// Portions of this file are additionally subject to the following
// license and copyright.
//
// Copyright 2015 The Cockroach Authors.
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

// This code was derived from https://github.com/youtube/vitess.

package tree

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/errors"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

// SelectStatement represents any SELECT statement.
type SelectStatement interface {
	Statement
	selectStatement()
}

func (*ParenSelect) selectStatement()  {}
func (*SelectClause) selectStatement() {}
func (*UnionClause) selectStatement()  {}
func (*ValuesClause) selectStatement() {}

// Select represents a SelectStatement with an ORDER and/or LIMIT.
type Select struct {
	With    *With
	Select  SelectStatement
	OrderBy OrderBy
	Limit   *Limit
}

// Format implements the NodeFormatter interface.
func (node *Select) Format(ctx *FmtCtx) {
	ctx.FormatNode(node.With)
	ctx.FormatNode(node.Select)
	if len(node.OrderBy) > 0 {
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.OrderBy)
	}
	if node.Limit != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Limit)
	}
}

// ParenSelect represents a parenthesized SELECT/UNION/VALUES statement.
type ParenSelect struct {
	Select *Select
}

// Format implements the NodeFormatter interface.
func (node *ParenSelect) Format(ctx *FmtCtx) {
	ctx.WriteByte('(')
	ctx.FormatNode(node.Select)
	ctx.WriteByte(')')
}

// SelectClause represents a SELECT statement.
type SelectClause struct {
	Distinct    bool
	DistinctOn  DistinctOn
	Exprs       SelectExprs
	From        *From
	Where       *Where
	GroupBy     GroupBy
	Having      *Where
	Window      Window
	TableSelect bool
}

// Format implements the NodeFormatter interface.
func (node *SelectClause) Format(ctx *FmtCtx) {
	if node.TableSelect {
		ctx.WriteString("TABLE ")
		ctx.FormatNode(node.From.Tables[0])
	} else {
		ctx.WriteString("SELECT ")
		if node.Distinct {
			if node.DistinctOn != nil {
				ctx.FormatNode(&node.DistinctOn)
				ctx.WriteByte(' ')
			} else {
				ctx.WriteString("DISTINCT ")
			}
		}
		ctx.FormatNode(&node.Exprs)
		if len(node.From.Tables) > 0 {
			ctx.WriteByte(' ')
			ctx.FormatNode(node.From)
		}
		if node.Where != nil {
			ctx.WriteByte(' ')
			ctx.FormatNode(node.Where)
		}
		if len(node.GroupBy) > 0 {
			ctx.WriteByte(' ')
			ctx.FormatNode(&node.GroupBy)
		}
		if node.Having != nil {
			ctx.WriteByte(' ')
			ctx.FormatNode(node.Having)
		}
		if len(node.Window) > 0 {
			ctx.WriteByte(' ')
			ctx.FormatNode(&node.Window)
		}
	}
}

// SelectExprs represents SELECT expressions.
type SelectExprs []SelectExpr

// Format implements the NodeFormatter interface.
func (node *SelectExprs) Format(ctx *FmtCtx) {
	for i := range *node {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&(*node)[i])
	}
}

// SelectExpr represents a SELECT expression.
type SelectExpr struct {
	Expr Expr
	As   UnrestrictedName
}

// NormalizeTopLevelVarName preemptively expands any UnresolvedName at
// the top level of the expression into a VarName. This is meant
// to catch stars so that sql.checkRenderStar() can see it prior to
// other expression transformations.
func (node *SelectExpr) NormalizeTopLevelVarName() error {
	if vBase, ok := node.Expr.(VarName); ok {
		v, err := vBase.NormalizeVarName()
		if err != nil {
			return err
		}
		node.Expr = v
	}
	return nil
}

// StarSelectExpr is a convenience function that represents an unqualified "*"
// in a select expression.
func StarSelectExpr() SelectExpr {
	return SelectExpr{Expr: StarExpr()}
}

// Format implements the NodeFormatter interface.
func (node *SelectExpr) Format(ctx *FmtCtx) {
	ctx.FormatNode(node.Expr)
	if node.As != "" {
		ctx.WriteString(" AS ")
		ctx.FormatNode(&node.As)
	}
}

// AliasClause represents an alias, optionally with a column list:
// "AS name" or "AS name(col1, col2)".
type AliasClause struct {
	Alias Name
	Cols  NameList
}

// Format implements the NodeFormatter interface.
func (a *AliasClause) Format(ctx *FmtCtx) {
	ctx.FormatNode(&a.Alias)
	if len(a.Cols) != 0 {
		// Format as "alias (col1, col2, ...)".
		ctx.WriteString(" (")
		ctx.FormatNode(&a.Cols)
		ctx.WriteByte(')')
	}
}

// AsOfClause represents an as of time.
type AsOfClause struct {
	Expr Expr
}

// Format implements the NodeFormatter interface.
func (a *AsOfClause) Format(ctx *FmtCtx) {
	ctx.WriteString("AS OF SYSTEM TIME ")
	ctx.FormatNode(a.Expr)
}

// From represents a FROM clause.
type From struct {
	Tables TableExprs
	AsOf   AsOfClause
}

// Format implements the NodeFormatter interface.
func (node *From) Format(ctx *FmtCtx) {
	ctx.WriteString("FROM ")
	ctx.FormatNode(&node.Tables)
	if node.AsOf.Expr != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.AsOf)
	}
}

// TableExprs represents a list of table expressions.
type TableExprs []TableExpr

// Format implements the NodeFormatter interface.
func (node *TableExprs) Format(ctx *FmtCtx) {
	prefix := ""
	for _, n := range *node {
		ctx.WriteString(prefix)
		ctx.FormatNode(n)
		prefix = ", "
	}
}

// TableExpr represents a table expression.
type TableExpr interface {
	NodeFormatter
	tableExpr()
}

func (*AliasedTableExpr) tableExpr() {}
func (*ParenTableExpr) tableExpr()   {}
func (*JoinTableExpr) tableExpr()    {}
func (*RowsFromExpr) tableExpr()     {}
func (*Subquery) tableExpr()         {}
func (*StatementSource) tableExpr()  {}

// StatementSource encapsulates one of the other statements as a data source.
type StatementSource struct {
	Statement Statement
}

// Format implements the NodeFormatter interface.
func (node *StatementSource) Format(ctx *FmtCtx) {
	ctx.WriteByte('[')
	ctx.FormatNode(node.Statement)
	ctx.WriteByte(']')
}

// IndexID is a custom type for IndexDescriptor IDs.
type IndexID uint32

// IndexFlags represents "@<index_name|index_id>" or "@{param[,param]}" where
// param is one of:
//  - FORCE_INDEX=<index_name|index_id>
//  - ASC / DESC
//  - NO_INDEX_JOIN
//  - IGNORE_FOREIGN_KEYS
// It is used optionally after a table name in SELECT statements.
type IndexFlags struct {
	Index   UnrestrictedName
	IndexID IndexID
	// Direction of the scan, if provided. Can only be set if
	// one of Index or IndexID is set.
	Direction Direction
	// NoIndexJoin cannot be specified together with an index.
	NoIndexJoin bool
	// IgnoreForeignKeys disables optimizations based on outbound foreign key
	// references from this table. This is useful in particular for scrub queries
	// used to verify the consistency of foreign key relations.
	IgnoreForeignKeys bool
}

// ForceIndex returns true if a forced index was specified, either using a name
// or an IndexID.
func (ih *IndexFlags) ForceIndex() bool {
	return ih.Index != "" || ih.IndexID != 0
}

// CombineWith combines two IndexFlags structures, returning an error if they
// conflict with one another.
func (ih *IndexFlags) CombineWith(other *IndexFlags) error {
	if ih.NoIndexJoin && other.NoIndexJoin {
		return errors.New("NO_INDEX_JOIN specified multiple times")
	}
	if ih.IgnoreForeignKeys && other.IgnoreForeignKeys {
		return errors.New("IGNORE_FOREIGN_KEYS specified multiple times")
	}
	result := *ih
	result.NoIndexJoin = ih.NoIndexJoin || other.NoIndexJoin
	result.IgnoreForeignKeys = ih.IgnoreForeignKeys || other.IgnoreForeignKeys

	if other.Direction != 0 {
		if ih.Direction != 0 {
			return errors.New("ASC/DESC specified multiple times")
		}
		result.Direction = other.Direction
	}

	if other.ForceIndex() {
		if ih.ForceIndex() {
			return errors.New("FORCE_INDEX specified multiple times")
		}
		result.Index = other.Index
		result.IndexID = other.IndexID
	}

	// We only set at the end to avoid a partially changed structure in one of the
	// error cases above.
	*ih = result
	return nil
}

// Check verifies if the flags are valid:
//  - ascending/descending is not specified without an index;
//  - no_index_join isn't specified with an index.
func (ih *IndexFlags) Check() error {
	if ih.NoIndexJoin && ih.ForceIndex() {
		return errors.New("FORCE_INDEX cannot be specified in conjunction with NO_INDEX_JOIN")
	}
	if ih.Direction != 0 && !ih.ForceIndex() {
		return errors.New("ASC/DESC must be specified in conjunction with an index")
	}
	return nil
}

// Format implements the NodeFormatter interface.
func (ih *IndexFlags) Format(ctx *FmtCtx) {
	ctx.WriteByte('@')
	if !ih.NoIndexJoin && !ih.IgnoreForeignKeys && ih.Direction == 0 {
		if ih.Index != "" {
			ctx.FormatNode(&ih.Index)
		} else {
			ctx.Printf("[%d]", ih.IndexID)
		}
	} else {
		ctx.WriteByte('{')
		var sep func()
		sep = func() {
			sep = func() { ctx.WriteByte(',') }
		}
		if ih.Index != "" || ih.IndexID != 0 {
			sep()
			ctx.WriteString("FORCE_INDEX=")
			if ih.Index != "" {
				ctx.FormatNode(&ih.Index)
			} else {
				ctx.Printf("[%d]", ih.IndexID)
			}

			if ih.Direction != 0 {
				ctx.Printf(",%s", ih.Direction)
			}
		}
		if ih.NoIndexJoin {
			sep()
			ctx.WriteString("NO_INDEX_JOIN")
		}

		if ih.IgnoreForeignKeys {
			sep()
			ctx.WriteString("IGNORE_FOREIGN_KEYS")
		}
		ctx.WriteString("}")
	}
}

// AliasedTableExpr represents a table expression coupled with an optional
// alias.
type AliasedTableExpr struct {
	Expr       TableExpr
	IndexFlags *IndexFlags
	Ordinality bool
	Lateral    bool
	As         AliasClause
}

// Format implements the NodeFormatter interface.
func (node *AliasedTableExpr) Format(ctx *FmtCtx) {
	if node.Lateral {
		ctx.WriteString("LATERAL ")
	}
	ctx.FormatNode(node.Expr)
	if node.IndexFlags != nil {
		ctx.FormatNode(node.IndexFlags)
	}
	if node.Ordinality {
		ctx.WriteString(" WITH ORDINALITY")
	}
	if node.As.Alias != "" {
		ctx.WriteString(" AS ")
		ctx.FormatNode(&node.As)
	}
}

// ParenTableExpr represents a parenthesized TableExpr.
type ParenTableExpr struct {
	Expr TableExpr
}

// Format implements the NodeFormatter interface.
func (node *ParenTableExpr) Format(ctx *FmtCtx) {
	ctx.WriteByte('(')
	ctx.FormatNode(node.Expr)
	ctx.WriteByte(')')
}

// StripTableParens strips any parentheses surrounding a selection clause.
func StripTableParens(expr TableExpr) TableExpr {
	if p, ok := expr.(*ParenTableExpr); ok {
		return StripTableParens(p.Expr)
	}
	return expr
}

// JoinTableExpr represents a TableExpr that's a JOIN operation.
type JoinTableExpr struct {
	JoinType string
	Left     TableExpr
	Right    TableExpr
	Cond     JoinCond
	Hint     string
}

// JoinTableExpr.Join
const (
	AstFull  = "FULL"
	AstLeft  = "LEFT"
	AstRight = "RIGHT"
	AstCross = "CROSS"
	AstInner = "INNER"
)

// JoinTableExpr.Hint
const (
	AstHash   = "HASH"
	AstLookup = "LOOKUP"
	AstMerge  = "MERGE"
)

// Format implements the NodeFormatter interface.
func (node *JoinTableExpr) Format(ctx *FmtCtx) {
	ctx.FormatNode(node.Left)
	ctx.WriteByte(' ')
	if _, isNatural := node.Cond.(NaturalJoinCond); isNatural {
		// Natural joins have a different syntax: "<a> NATURAL <join_type> <b>"
		ctx.FormatNode(node.Cond)
		ctx.WriteByte(' ')
		if node.JoinType != "" {
			ctx.WriteString(node.JoinType)
			ctx.WriteByte(' ')
			if node.Hint != "" {
				ctx.WriteString(node.Hint)
				ctx.WriteByte(' ')
			}
		}
		ctx.WriteString("JOIN ")
		ctx.FormatNode(node.Right)
	} else {
		// General syntax: "<a> <join_type> [<join_hint>] JOIN <b> <condition>"
		if node.JoinType != "" {
			ctx.WriteString(node.JoinType)
			ctx.WriteByte(' ')
			if node.Hint != "" {
				ctx.WriteString(node.Hint)
				ctx.WriteByte(' ')
			}
		}
		ctx.WriteString("JOIN ")
		ctx.FormatNode(node.Right)
		if node.Cond != nil {
			ctx.WriteByte(' ')
			ctx.FormatNode(node.Cond)
		}
	}
}

// JoinCond represents a join condition.
type JoinCond interface {
	NodeFormatter
	joinCond()
}

func (NaturalJoinCond) joinCond() {}
func (*OnJoinCond) joinCond()     {}
func (*UsingJoinCond) joinCond()  {}

// NaturalJoinCond represents a NATURAL join condition
type NaturalJoinCond struct{}

// Format implements the NodeFormatter interface.
func (NaturalJoinCond) Format(ctx *FmtCtx) {
	ctx.WriteString("NATURAL")
}

// OnJoinCond represents an ON join condition.
type OnJoinCond struct {
	Expr Expr
}

// Format implements the NodeFormatter interface.
func (node *OnJoinCond) Format(ctx *FmtCtx) {
	ctx.WriteString("ON ")
	ctx.FormatNode(node.Expr)
}

// UsingJoinCond represents a USING join condition.
type UsingJoinCond struct {
	Cols NameList
}

// Format implements the NodeFormatter interface.
func (node *UsingJoinCond) Format(ctx *FmtCtx) {
	ctx.WriteString("USING (")
	ctx.FormatNode(&node.Cols)
	ctx.WriteByte(')')
}

// Where represents a WHERE or HAVING clause.
type Where struct {
	Type string
	Expr Expr
}

// Where.Type
const (
	AstWhere  = "WHERE"
	AstHaving = "HAVING"
)

// NewWhere creates a WHERE or HAVING clause out of an Expr. If the expression
// is nil, it returns nil.
func NewWhere(typ string, expr Expr) *Where {
	if expr == nil {
		return nil
	}
	return &Where{Type: typ, Expr: expr}
}

// Format implements the NodeFormatter interface.
func (node *Where) Format(ctx *FmtCtx) {
	ctx.WriteString(node.Type)
	ctx.WriteByte(' ')
	ctx.FormatNode(node.Expr)
}

// GroupBy represents a GROUP BY clause.
type GroupBy []Expr

// Format implements the NodeFormatter interface.
func (node *GroupBy) Format(ctx *FmtCtx) {
	prefix := "GROUP BY "
	for _, n := range *node {
		ctx.WriteString(prefix)
		ctx.FormatNode(n)
		prefix = ", "
	}
}

// DistinctOn represents a DISTINCT ON clause.
type DistinctOn []Expr

// Format implements the NodeFormatter interface.
func (node *DistinctOn) Format(ctx *FmtCtx) {
	ctx.WriteString("DISTINCT ON (")
	ctx.FormatNode((*Exprs)(node))
	ctx.WriteByte(')')
}

// OrderBy represents an ORDER By clause.
type OrderBy []*Order

// Format implements the NodeFormatter interface.
func (node *OrderBy) Format(ctx *FmtCtx) {
	prefix := "ORDER BY "
	for _, n := range *node {
		ctx.WriteString(prefix)
		ctx.FormatNode(n)
		prefix = ", "
	}
}

// Direction for ordering results.
type Direction int8

// Direction values.
const (
	DefaultDirection Direction = iota
	Ascending
	Descending
)

var directionName = [...]string{
	DefaultDirection: "",
	Ascending:        "ASC",
	Descending:       "DESC",
}

func (d Direction) String() string {
	if d < 0 || d > Direction(len(directionName)-1) {
		return fmt.Sprintf("Direction(%d)", d)
	}
	return directionName[d]
}

// OrderType indicates which type of expression is used in ORDER BY.
type OrderType int

const (
	// OrderByColumn is the regular "by expression/column" ORDER BY specification.
	OrderByColumn OrderType = iota
	// OrderByIndex enables the user to specify a given index' columns implicitly.
	OrderByIndex
)

// Order represents an ordering expression.
type Order struct {
	OrderType OrderType
	Expr      Expr
	Direction Direction
	// Table/Index replaces Expr when OrderType = OrderByIndex.
	Table TableName
	// If Index is empty, then the order should use the primary key.
	Index UnrestrictedName
}

// Format implements the NodeFormatter interface.
func (node *Order) Format(ctx *FmtCtx) {
	if node.OrderType == OrderByColumn {
		ctx.FormatNode(node.Expr)
	} else {
		if node.Index == "" {
			ctx.WriteString("PRIMARY KEY ")
			ctx.FormatNode(&node.Table)
		} else {
			ctx.WriteString("INDEX ")
			ctx.FormatNode(&node.Table)
			ctx.WriteByte('@')
			ctx.FormatNode(&node.Index)
		}
	}
	if node.Direction != DefaultDirection {
		ctx.WriteByte(' ')
		ctx.WriteString(node.Direction.String())
	}
}

// Limit represents a LIMIT clause.
type Limit struct {
	Offset, Count Expr
}

// Format implements the NodeFormatter interface.
func (node *Limit) Format(ctx *FmtCtx) {
	needSpace := false
	if node.Count != nil {
		ctx.WriteString("LIMIT ")
		ctx.FormatNode(node.Count)
		needSpace = true
	}
	if node.Offset != nil {
		if needSpace {
			ctx.WriteByte(' ')
		}
		ctx.WriteString("OFFSET ")
		ctx.FormatNode(node.Offset)
	}
}

// RowsFromExpr represents a ROWS FROM(...) expression.
type RowsFromExpr struct {
	Items Exprs
}

// Format implements the NodeFormatter interface.
func (node *RowsFromExpr) Format(ctx *FmtCtx) {
	ctx.WriteString("ROWS FROM (")
	ctx.FormatNode(&node.Items)
	ctx.WriteByte(')')
}

// Window represents a WINDOW clause.
type Window []*WindowDef

// Format implements the NodeFormatter interface.
func (node *Window) Format(ctx *FmtCtx) {
	prefix := "WINDOW "
	for _, n := range *node {
		ctx.WriteString(prefix)
		ctx.FormatNode(&n.Name)
		ctx.WriteString(" AS ")
		ctx.FormatNode(n)
		prefix = ", "
	}
}

// WindowDef represents a single window definition expression.
type WindowDef struct {
	Name       Name
	RefName    Name
	Partitions Exprs
	OrderBy    OrderBy
	Frame      *WindowFrame
}

// Format implements the NodeFormatter interface.
func (node *WindowDef) Format(ctx *FmtCtx) {
	ctx.WriteByte('(')
	needSpaceSeparator := false
	if node.RefName != "" {
		ctx.FormatNode(&node.RefName)
		needSpaceSeparator = true
	}
	if len(node.Partitions) > 0 {
		if needSpaceSeparator {
			ctx.WriteByte(' ')
		}
		ctx.WriteString("PARTITION BY ")
		ctx.FormatNode(&node.Partitions)
		needSpaceSeparator = true
	}
	if len(node.OrderBy) > 0 {
		if needSpaceSeparator {
			ctx.WriteByte(' ')
		}
		ctx.FormatNode(&node.OrderBy)
		needSpaceSeparator = true
	}
	if node.Frame != nil {
		if needSpaceSeparator {
			ctx.WriteByte(' ')
		}
		ctx.FormatNode(node.Frame)
	}
	ctx.WriteRune(')')
}

// WindowFrameMode indicates which mode of framing is used.
type WindowFrameMode int

const (
	// RANGE is the mode of specifying frame in terms of logical range (e.g. 100 units cheaper).
	RANGE WindowFrameMode = iota
	// ROWS is the mode of specifying frame in terms of physical offsets (e.g. 1 row before etc).
	ROWS
	// GROUPS is the mode of specifying frame in terms of peer groups.
	GROUPS
)

// WindowFrameBoundType indicates which type of boundary is used.
type WindowFrameBoundType int

const (
	// UnboundedPreceding represents UNBOUNDED PRECEDING type of boundary.
	UnboundedPreceding WindowFrameBoundType = iota
	// OffsetPreceding represents 'value' PRECEDING type of boundary.
	OffsetPreceding
	// CurrentRow represents CURRENT ROW type of boundary.
	CurrentRow
	// OffsetFollowing represents 'value' FOLLOWING type of boundary.
	OffsetFollowing
	// UnboundedFollowing represents UNBOUNDED FOLLOWING type of boundary.
	UnboundedFollowing
)

// WindowFrameBound specifies the offset and the type of boundary.
type WindowFrameBound struct {
	BoundType  WindowFrameBoundType
	OffsetExpr Expr
}

// HasOffset returns whether node contains an offset.
func (node *WindowFrameBound) HasOffset() bool {
	return node.BoundType == OffsetPreceding || node.BoundType == OffsetFollowing
}

// WindowFrameBounds specifies boundaries of the window frame.
// The row at StartBound is included whereas the row at EndBound is not.
type WindowFrameBounds struct {
	StartBound *WindowFrameBound
	EndBound   *WindowFrameBound
}

// HasOffset returns whether node contains an offset in either of the bounds.
func (node *WindowFrameBounds) HasOffset() bool {
	return node.StartBound.HasOffset() || (node.EndBound != nil && node.EndBound.HasOffset())
}

// WindowFrame represents static state of window frame over which calculations are made.
type WindowFrame struct {
	Mode   WindowFrameMode   // the mode of framing being used
	Bounds WindowFrameBounds // the bounds of the frame
}

// Format implements the NodeFormatter interface.
func (node *WindowFrameBound) Format(ctx *FmtCtx) {
	switch node.BoundType {
	case UnboundedPreceding:
		ctx.WriteString("UNBOUNDED PRECEDING")
	case OffsetPreceding:
		ctx.FormatNode(node.OffsetExpr)
		ctx.WriteString(" PRECEDING")
	case CurrentRow:
		ctx.WriteString("CURRENT ROW")
	case OffsetFollowing:
		ctx.FormatNode(node.OffsetExpr)
		ctx.WriteString(" FOLLOWING")
	case UnboundedFollowing:
		ctx.WriteString("UNBOUNDED FOLLOWING")
	default:
		panic(errors.AssertionFailedf("unhandled case: %d", log.Safe(node.BoundType)))
	}
}

// Format implements the NodeFormatter interface.
func (node *WindowFrame) Format(ctx *FmtCtx) {
	switch node.Mode {
	case RANGE:
		ctx.WriteString("RANGE ")
	case ROWS:
		ctx.WriteString("ROWS ")
	case GROUPS:
		ctx.WriteString("GROUPS ")
	default:
		panic(errors.AssertionFailedf("unhandled case: %d", log.Safe(node.Mode)))
	}
	if node.Bounds.EndBound != nil {
		ctx.WriteString("BETWEEN ")
		ctx.FormatNode(node.Bounds.StartBound)
		ctx.WriteString(" AND ")
		ctx.FormatNode(node.Bounds.EndBound)
	} else {
		ctx.FormatNode(node.Bounds.StartBound)
	}
}
