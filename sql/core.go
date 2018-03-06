package sql

import (
	"gopkg.in/src-d/go-errors.v1"
)

// Nameable is something that has a name.
type Nameable interface {
	// Name returns the name.
	Name() string
}

// Resolvable is something that can be resolved or not.
type Resolvable interface {
	// Resolved returns whether the node is resolved.
	Resolved() bool
}

// Transformable is a node which can be transformed.
type Transformable interface {
	// TransformUp transforms all nodes and returns the result of this transformation.
	TransformUp(func(Node) (Node, error)) (Node, error)
	// TransformExpressionsUp transforms all expressions inside the node and all its
	// children and returns a node with the result of the transformations.
	TransformExpressionsUp(func(Expression) (Expression, error)) (Node, error)
}

// Expression is a combination of one or more SQL expressions.
type Expression interface {
	Resolvable
	// Type returns the expression type.
	Type() Type
	// Name returns the expression name.
	Name() string
	// IsNullable returns whether the expression can be null.
	IsNullable() bool
	// Eval evaluates the given row and returns a result.
	Eval(Session, Row) (interface{}, error)
	// TransformUp transforms the expression and all its children with the
	// given transform function.
	TransformUp(func(Expression) (Expression, error)) (Expression, error)
}

// Aggregation implements an aggregation expression, where an
// aggregation buffer is created for each grouping (NewBuffer) and rows in the
// grouping are fed to the buffer (Update). Multiple buffers can be merged
// (Merge), making partial aggregations possible.
// Note that Eval must be called with the final aggregation buffer in order to
// get the final result.
type Aggregation interface {
	Expression
	// NewBuffer creates a new aggregation buffer and returns it as a Row.
	NewBuffer() Row
	// Update updates the given buffer with the given row.
	Update(session Session, buffer, row Row) error
	// Merge merges a partial buffer into a global one.
	Merge(session Session, buffer, partial Row) error
}

// Node is a node in the execution plan tree.
type Node interface {
	Resolvable
	Transformable
	// Schema of the node.
	Schema() Schema
	// Children nodes.
	Children() []Node
	// RowIter produces a row iterator from this node.
	RowIter(Session) (RowIter, error)
}

// Table represents a SQL table.
type Table interface {
	Nameable
	Node
}

// PushdownProjectionTable is a table that can produce a specific RowIter
// that's more optimized given the columns that are projected.
type PushdownProjectionTable interface {
	Table
	// WithProject replaces the RowIter method of the table and returns a new
	// row iterator given the column names that are projected.
	WithProject(session Session, colNames []string) (RowIter, error)
}

// PushdownProjectionAndFiltersTable is a table that can produce a specific
// RowIter that's more optimized given the columns that are projected and
// the filters for this table.
type PushdownProjectionAndFiltersTable interface {
	Table
	// HandledFilters returns the subset of filters that can be handled by this
	// table.
	HandledFilters(filters []Expression) []Expression
	// WithProjectAndFilters replaces the RowIter method of the table and
	// return a new row iterator given the column names that are projected
	// and the filters applied to this table.
	WithProjectAndFilters(session Session, columns, filters []Expression) (RowIter, error)
}

// Inserter allow rows to be inserted in them.
type Inserter interface {
	// Insert the given row.
	Insert(row Row) error
}

// Database represents the database.
type Database interface {
	Nameable
	// Tables returns the information of all tables.
	Tables() map[string]Table
}

// Alterable should be implemented by databases that can handle DDL statements
type Alterable interface {
	Create(name string, schema Schema) error
}

// ErrInvalidType is thrown when there is an unexpected type at some part of
// the execution tree.
var ErrInvalidType = errors.NewKind("invalid type: %s")
