package database

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// There is no wrapper so we have to check the SQL state that bubbled up
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func CheckNoRow(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func IsForeignKeyViolations(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
