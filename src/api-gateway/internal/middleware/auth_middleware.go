package middleware

import (
	"IDIG4110/shared/apperror"
	"IDIG4110/shared/httperror"
	"context"
	"crypto/ed25519"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const prefix = "Bearer "

type userIDContextKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(string)
	return userID, ok
}

func AuthMiddleware(publicKey ed25519.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				httperror.HandleAppError(w, apperror.UnathorizedRequestError("Missing auth header"))
				return
			}

			if !strings.HasPrefix(authHeader, prefix) {
				httperror.HandleAppError(w, apperror.UnathorizedRequestError("Expected the prefix: "+prefix))
				return
			}

			rawToken := strings.TrimPrefix(authHeader, prefix)

			token, err := jwt.Parse(
				rawToken,
				func(token *jwt.Token) (any, error) {
					if token.Method != jwt.SigningMethodEdDSA {
						return nil, fmt.Errorf("Unexpected signing method")
					}

					return publicKey, nil
				},
				jwt.WithIssuer("auth-service"),
				jwt.WithAudience("smarthome"),
				jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
			)
			if err != nil || !token.Valid {
				httperror.HandleAppError(w, apperror.UnathorizedRequestError("malformed or expired token"))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				httperror.HandleAppError(w, apperror.UnathorizedRequestError())
				return
			}

			userID, err := claims.GetSubject()
			if err != nil || userID == "" {
				apperror.UnathorizedRequestError()
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDContextKey{},
				userID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
