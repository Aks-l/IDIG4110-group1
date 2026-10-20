package auth

import (
	"IDIG4110/auth-service/internal"
	"IDIG4110/shared/apperror"
	"IDIG4110/shared/dto"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

type authRepository interface {
	CreateUser(ctx context.Context, user User) error
	FindUserByEmail(ctx context.Context, email string) (*LoginUser, error)
	CreateSession(ctx context.Context, sesh Session) error
	GetSessionByHash(ctx context.Context, refreshHash string) (*Session, error)
	ReplaceRefreshHash(ctx context.Context, newRefreshHash, oldRefreshHash, userID string) error
	RevokeSessionByHash(ctx context.Context, refreshHash string) error
}

type tokenSigner interface {
	CreateAccessToken(userID string) (string, error)
}

type AuthService struct {
	repo   authRepository
	signer tokenSigner
}

func NewAuthService(
	userRepo authRepository,
	signer tokenSigner,
) *AuthService {
	return &AuthService{
		repo:   userRepo,
		signer: signer,
	}
}

func hashPassword(password string) (string, error) {
	salt, err := internal.GenerateSalt()
	// The package itself says this should realistically never fail
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		threads,
		keyLength,
	)

	return fmt.Sprintf(
		hashStructure,
		argon2.Version,
		memory,
		iterations,
		threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (s *AuthService) RegisterNewAccount(ctx context.Context, userDTO dto.UserDTO) error {
	hashedPassword, err := hashPassword(userDTO.Password)
	if err != nil {
		return err
	}

	return s.repo.CreateUser(ctx, FromCreateDTO(userDTO, hashedPassword))
}

func verifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {

	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}

	var parsedMemory, parsedIterations uint32
	var parsedThreads uint8
	if _, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&parsedMemory,
		&parsedIterations,
		&parsedThreads,
	); err != nil {

	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, apperror.InternalServerError()
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, apperror.InternalServerError()
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		parsedIterations,
		parsedMemory,
		parsedThreads,
		uint32(len(expectedHash)),
	)

	same := subtle.ConstantTimeCompare(expectedHash, actualHash) == 1

	return same, nil
}

func createRefreshToken() (string, string, error) {
	token := make([]byte, 32)

	if _, err := rand.Read(token); err != nil {
		return "", "", err
	}

	raw := base64.RawURLEncoding.EncodeToString(token)

	hash := sha256.Sum256([]byte(raw))
	hashed := hex.EncodeToString(hash[:])

	return raw, hashed, nil
}

func (s *AuthService) Login(ctx context.Context, userDTO dto.UserDTO) (*dto.TokenResponseDTO, error) {
	user, err := s.repo.FindUserByEmail(ctx, userDTO.Email)

	valid, err := verifyPassword(userDTO.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil,
			apperror.CreateNewError(
				apperror.CodeInvalidInput,
				"Wrong password",
			)
	}

	accessToken, err := s.signer.CreateAccessToken(user.ID)
	refreshToken, refresHash, err := createRefreshToken()

	session := Session{
		UserID:           user.ID,
		RefreshTokenHash: refresHash,
		ExpiresAt:        time.Now().Add(refreshTokenDuration),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return &dto.TokenResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresIn:    int(accessTokenDuration.Seconds()),
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refresh dto.RefreshTokenDTO) (*dto.TokenResponseDTO, error) {
	hash := sha256.Sum256([]byte(refresh.RefreshToken))
	hashed := hex.EncodeToString(hash[:])

	session, err := s.repo.GetSessionByHash(ctx, hashed)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.signer.CreateAccessToken(session.UserID)
	if err != nil {
		return nil, err
	}

	refreshToken, newRefreshHash, err := createRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.ReplaceRefreshHash(
		ctx,
		newRefreshHash,
		session.RefreshTokenHash,
		session.UserID,
	); err != nil {
		return nil, err
	}

	return &dto.TokenResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresIn:    int(accessTokenDuration.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refresh dto.RefreshTokenDTO) error {
	hash := sha256.Sum256([]byte(refresh.RefreshToken))
	hashed := hex.EncodeToString(hash[:])

	return s.repo.RevokeSessionByHash(ctx, hashed)
}
