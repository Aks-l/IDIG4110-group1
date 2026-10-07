// Package pgerrors classifies PostgreSQL errors returned through pgx so
// each service can map constraint violations onto its own error contract
// without repeating the error-code table.
package pgerrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes; see
// https://www.postgresql.org/docs/current/errcodes-appendix.html.
const (
	CodeUniqueViolation     = "23505"
	CodeForeignKeyViolation = "23503"
	CodeNotNullViolation    = "23502"
	CodeCheckViolation      = "23514"
	CodeStringValueTooLong  = "22001"
)

// Code returns the PostgreSQL error code carried by err, or "" when err is
// not a PostgreSQL error.
func Code(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// Message returns the PostgreSQL server's error message, or "" when err is
// not a PostgreSQL error.
func Message(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Message
	}
	return ""
}

// IsUniqueViolation reports a 23505, e.g. a duplicate key.
func IsUniqueViolation(err error) bool {
	return Code(err) == CodeUniqueViolation
}

// IsForeignKeyViolation reports a 23503, e.g. a row that references a
// missing parent.
func IsForeignKeyViolation(err error) bool {
	return Code(err) == CodeForeignKeyViolation
}

// IsNotNullViolation reports a 23502.
func IsNotNullViolation(err error) bool {
	return Code(err) == CodeNotNullViolation
}

// IsCheckViolation reports a 23514, e.g. a value outside a CHECK domain.
func IsCheckViolation(err error) bool {
	return Code(err) == CodeCheckViolation
}

// IsStringValueTooLong reports a 22001.
func IsStringValueTooLong(err error) bool {
	return Code(err) == CodeStringValueTooLong
}
