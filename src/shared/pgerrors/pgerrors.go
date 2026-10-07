// Package pgerrors: classifies postgres errors for constraint mapping
package pgerrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes, see errcodes appendix in postgres docs
const (
	CodeUniqueViolation     = "23505"
	CodeForeignKeyViolation = "23503"
	CodeNotNullViolation    = "23502"
	CodeCheckViolation      = "23514"
	CodeStringValueTooLong  = "22001"
)

// Extracts postgres error code from err
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - Error code, empty when not a postgres error
func Code(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// Extracts postgres error message from err
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - Server message, empty when not a postgres error
func Message(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Message
	}
	return ""
}

// Reports duplicate key violation (23505)
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - True on unique violation
func IsUniqueViolation(err error) bool {
	return Code(err) == CodeUniqueViolation
}

// Reports missing reference violation (23503)
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - True on foreign key violation
func IsForeignKeyViolation(err error) bool {
	return Code(err) == CodeForeignKeyViolation
}

// Reports not null violation (23502)
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - True on not null violation
func IsNotNullViolation(err error) bool {
	return Code(err) == CodeNotNullViolation
}

// Reports check violation (23514)
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - True on check violation
func IsCheckViolation(err error) bool {
	return Code(err) == CodeCheckViolation
}

// Reports too long string value (22001)
//
// # Inputs:
//
//   - err [error] error to inspect
//
// # Returns:
//
//   - True on string value too long
func IsStringValueTooLong(err error) bool {
	return Code(err) == CodeStringValueTooLong
}
