package user

import (
	"context"
	"log"

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

	return err
}
