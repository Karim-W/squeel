//
//  pipeline.go
//  squeel
//
//  Created by karim-w on 10/07/2025.
//

package squeel

import (
	"context"
	"database/sql"
)

type pipeline struct {
	queryer     Queryer
	middlewares []Middleware
}

// ExecContext follows the native sql package's ExecContext signature.
// It executes a query that doesn't return rows, such as an INSERT, UPDATE, or DELETE.
// It returns the result of the query execution, which includes the number of rows affected.
// used the same way as the normal SQL package but its here in case you wanna add
// some middleware to the query execution process.
func (p *pipeline) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	stmt := &Statment{
		operation: _OPERATIONS.EXEC,
		Query:     query,
		Args:      args,
	}

	p.Run(stmt)

	return stmt.Result, stmt.error
}

// QueryContext follows the native sql package's QueryContext signature.
// It executes a query that returns **MULTIPLE** rows
// It returns the rows that were returned by the query.
// This method is used the same way as the normal SQL package but it's here in case you want to add
// some middleware to the query execution process.
func (p *pipeline) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	stmt := &Statment{
		operation: _OPERATIONS.QUERY,
		Query:     query,
		Args:      args,
	}

	p.Run(stmt)

	return stmt.Rows, stmt.error
}

// QueryRowContext follows the native sql package's QueryRowContext signature.
// It executes a query that returns a single row.
// It returns a *sql.Row that can be used to scan the result.
// This method is used the same way as the normal SQL package but it's here in case you want to add
// some middleware to the query execution process.
func (p *pipeline) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	stmt := &Statment{
		operation: _OPERATIONS.QUERY_ROW,
		Query:     query,
		Args:      args,
	}

	p.Run(stmt)

	return stmt.Row
}

var _ Pipeline = (*pipeline)(nil)

// Use adds middleware to the pipeline.
// Middleware functions are applied in the order they are added, allowing for
// flexible query processing. Each middleware can in theory modify the query or
// the execution context (but should it? is a completely diff question)
// to add logging, modify the query parameters, handle errors or perform
// any additional operations before or after the query execution.
func (p *pipeline) Use(middleware ...Middleware) {
	p.middlewares = append(p.middlewares, middleware...)
}

func (p *pipeline) Run(stmt *Statment) {
	index := 0

	var next func()
	next = func() {
		if index >= len(p.middlewares) {
			p.execute(stmt)
			return
		}
		mw := p.middlewares[index]
		index++
		stmt.next = next
		mw(stmt)
	}

	stmt.next = next
	stmt.next() // start chain
}

func (p *pipeline) execute(s *Statment) {
	switch s.operation {
	case _OPERATIONS.EXEC:
		s.Result, s.error = p.queryer.ExecContext(context.Background(), s.Query, s.Args...)
	case _OPERATIONS.QUERY:
		s.Rows, s.error = p.queryer.QueryContext(context.Background(), s.Query, s.Args...)
	case _OPERATIONS.QUERY_ROW:
		s.Row = p.queryer.QueryRowContext(context.Background(), s.Query, s.Args...)
	}
}
