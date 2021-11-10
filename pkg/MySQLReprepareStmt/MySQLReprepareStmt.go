package MySQLReprepareStmt

import (
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

type Stmt struct {
	stmt  *sql.Stmt
	query string
	db    *sql.DB
}

func IsRePrepareError(err error) bool {
	if err == nil {
		return false
	}
	if myErr, ok := err.(*mysql.MySQLError); ok {
		return myErr.Number == 1615
	}
	return false
}

func (s *Stmt) RePrepare() error {
	var err error
	if err = s.stmt.Close(); err != nil {
		return errors.Wrapf(err, "cannot close statement - %s", s.query)
	}
	if s.stmt, err = s.db.Prepare(s.query); err != nil {
		return errors.Wrapf(err, "cannot prepare statement - %s", s.query)
	}
	return nil
}

func (s *Stmt) RePrepareContext(ctx context.Context) error {
	var err error
	if err = s.stmt.Close(); err != nil {
		return errors.Wrapf(err, "cannot close statement - %s", s.query)
	}
	if s.stmt, err = s.db.PrepareContext(ctx, s.query); err != nil {
		return errors.Wrapf(err, "cannot prepare statement - %s", s.query)
	}
	return nil
}

func (s *Stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	result, err := s.stmt.ExecContext(ctx, args...)
	if IsRePrepareError(err) {
		if err := s.RePrepareContext(ctx); err != nil {
			return nil, err
		}
		result, err = s.stmt.ExecContext(ctx, args)
	}
	return result, err
}

func (s *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	result, err := s.stmt.Exec(args...)
	if IsRePrepareError(err) {
		if err := s.RePrepare(); err != nil {
			return nil, err
		}
		result, err = s.stmt.Exec(args)
	}
	return result, err
}

func (s *Stmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	rows, err := s.stmt.QueryContext(ctx, args...)
	if IsRePrepareError(err) {
		if err := s.RePrepareContext(ctx); err != nil {
			return nil, err
		}
		rows, err = s.stmt.QueryContext(ctx, args)
	}
	return rows, err
}

func (s *Stmt) Query(args ...interface{}) (*sql.Rows, error) {
	rows, err := s.stmt.Query(args...)
	if IsRePrepareError(err) {
		if err := s.RePrepare(); err != nil {
			return nil, err
		}
		rows, err = s.stmt.Query(args)
	}
	return rows, err
}

func (s *Stmt) QueryRow(args ...interface{}) *sql.Row {
	return s.stmt.QueryRow(args...)
}

func (s *Stmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	return s.stmt.QueryRowContext(ctx, args...)
}
