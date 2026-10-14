package auth

import (
	"time"
)

type User struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type LoginUser struct {
	ID           string `db:"id"`
	PasswordHash string `db:password_hash`
}

type Session struct {
	UserID           string    `db:"user_id"`
	RefreshTokenHash string    `db:"refresh_token"`
	ExpiresAt        time.Time `db:"expires_at"`
}
