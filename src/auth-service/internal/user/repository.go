package user

import (
	"IDIG4110/shared/apperror"
	database "IDIG4110/shared/db"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user User) error {
	const createUserQuery = `
INSERT INTO auth.users(
   email,
   password_hash
)
VALUES ($1, $2)
`

	_, err := r.db.Exec(
		ctx,
		createUserQuery,
		user.Email,
		user.PasswordHash,
	)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return apperror.CreateNewError(
				apperror.CodeEmailAlreadyUsed,
				"email alredy in use",
			)
		}
		return apperror.CreateNewError(
			apperror.CodeInternal,
			"internal server error",
		)
	}
	return nil
}
