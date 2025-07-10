// init.go
// squeel
//
// Created by karim-w on 10/07/2025.
package squeel

import (
	"context"
	"database/sql"
)

// Queryer is an interface that defines methods for executing SQL queries.
// It includes methods for executing commands that do not return rows, querying multiple rows,
// and querying a single row. This interface is designed to be implemented by types that
type Queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Transaction interface {
	Queryer
	Commit() error
	Rollback() error
}

type Connection interface {
	// Inheritance, great ...
	Queryer

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	PingContext(ctx context.Context) error
}

// Simple assertions to ensure that the types implement the interfaces
var (
	_ Connection  = (*sql.DB)(nil)
	_ Queryer     = (*sql.DB)(nil)
	_ Queryer     = (*sql.Conn)(nil)
	_ Queryer     = (*sql.Tx)(nil)
	_ Transaction = (*sql.Tx)(nil)
)

type Middleware func(
	statement *Statment,
)

type Pipeline interface {
	Use(middleware ...Middleware)
	Queryer
}

type SQL_OPERATION int64

var _OPERATIONS = struct {
	EXEC      SQL_OPERATION
	QUERY     SQL_OPERATION
	QUERY_ROW SQL_OPERATION
}{
	EXEC:      0,
	QUERY:     1,
	QUERY_ROW: 2,
}

type Statment struct {
	operation SQL_OPERATION
	Query     string
	Args      []any

	Rows   *sql.Rows
	Row    *sql.Row
	Result sql.Result

	error error
	next  func()
}

func (s *Statment) Next() {
	s.next()
}

func (s *Statment) Error() error {
	return s.error
}

func (s *Statment) OperationType() SQL_OPERATION {
	return s.operation
}

func NewPipeline(
	queryer Queryer,
) Pipeline {
	return &pipeline{
		middlewares: make([]Middleware, 0),
		queryer:     queryer,
	}
}
