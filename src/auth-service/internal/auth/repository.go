package auth

import (
	"IDIG4110/shared/apperror"
	database "IDIG4110/shared/db"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createUserQuery = `
INSERT INTO auth.users(
   email,
   password_hash
)
VALUES ($1, $2)
`

	findUserQuery = `
SELECT id, password_hash
FROM auth.users
WHERE email = $1
LIMIT 1
`
	createSessionQuery = `
INSERT INTO auth.sessions(
user_id,
refresh_token_hash,
expires_at
)
VALUES ($1, $2, $3)
`

	refreshTokenQuery = `
SELECT user_id, refresh_token_hash, expires_at
FROM auth.sessions
WHERE refresh_token_hash = $1
AND revoked_at IS NULL
AND expires_at > now()
LIMIT 1
`

	replaceRefreshHashQuery = `
UPDATE auth.sessions
SET refresh_token_hash = $1
WHERE refresh_token_hash = $2
AND user_id = $3
AND revoked_at IS NULL
AND expires_at > now()
`

	revokeSessionHash = `
UPDATE auth.sessions
SET revoked_at = now()
WHERE refresh_token_hash = $1
`
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateUser(ctx context.Context, user User) error {
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
		return apperror.InternalServerError()
	}
	return nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*LoginUser, error) {
	var fetchedUser LoginUser
	err := r.db.QueryRow(
		ctx,
		findUserQuery,
		email,
	).Scan(
		&fetchedUser.ID,

		&fetchedUser.PasswordHash,
	)
	if err != nil {
		if database.CheckNoRow(err) {
			return nil,
				apperror.CreateNewError(
					apperror.CodeNoContent,
					"No user by that name",
				)
		}
		return nil,
			apperror.InternalServerError()
	}

	return &fetchedUser, nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, sesh Session) error {
	_, err := r.db.Exec(
		ctx,
		createSessionQuery,
		sesh.UserID,
		sesh.RefreshTokenHash,
		sesh.ExpiresAt,
	)
	if err != nil {
		if database.IsForeignKeyViolations(err) {
			return apperror.CreateNewError(
				apperror.CodeNoContent,
				"No user by that name",
			)
		}
		return apperror.InternalServerError()
	}
	return nil
}

func (r *AuthRepository) GetSessionByHash(ctx context.Context, refreshHash string) (*Session, error) {
	var fetchedSession Session

	err := r.db.QueryRow(
		ctx,
		refreshTokenQuery,
		refreshHash,
	).Scan(
		&fetchedSession.UserID,
		&fetchedSession.RefreshTokenHash,
		&fetchedSession.ExpiresAt,
	)
	if err != nil {
		if database.CheckNoRow(err) {
			return nil,
				apperror.CreateNewError(
					apperror.CodeNoContent,
					"No valid session",
				)
		}
		return nil,
			apperror.InternalServerError()
	}
	return &fetchedSession, nil
}

func (r *AuthRepository) ReplaceRefreshHash(
	ctx context.Context,
	newRefreshHash, oldRefreshHash, userID string,
) error {
	result, err := r.db.Exec(
		ctx,
		replaceRefreshHashQuery,
		newRefreshHash,
		oldRefreshHash,
		userID,
	)
	if err != nil {
		return apperror.InternalServerError()
	}

	if result.RowsAffected() == 0 {
		return apperror.CreateNewError(
			apperror.CodeNoContent,
			"No valid session",
		)
	}

	return nil
}

func (r *AuthRepository) RevokeSessionByHash(ctx context.Context, refreshHash string) error {
	result, err := r.db.Exec(
		ctx,
		revokeSessionHash,
		refreshHash,
	)
	if err != nil {
		return apperror.InternalServerError()
	}

	if result.RowsAffected() == 0 {
		return apperror.CreateNewError(
			apperror.CodeNoContent,
			"No valid session",
		)
	}

	return nil
}
