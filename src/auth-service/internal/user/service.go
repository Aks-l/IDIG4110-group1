package user

import (
	"IDIG4110/auth-service/internal"
	"IDIG4110/shared/dto"
	"context"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	memory     = 64 * 1024
	iterations = 3
	threads    = 2
	keyLength  = 32
)

type userRepository interface {
	CreateUser(ctx context.Context, user User) error
}

type UserService struct {
	repo userRepository
}

func NewUserService(userRepo userRepository) *UserService {
	return &UserService{
		repo: userRepo,
	}
}

func hashPassword(password string) (string, error) {
	salt, err := internal.GenerateSalt()
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
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (s *UserService) RegisterNewAccount(ctx context.Context, userDTO dto.CreateUserDTO) error {
	hashedPassword, err := hashPassword(userDTO.Password)
	if err != nil {
		return err
	}

	s.repo.CreateUser(ctx, FromCreateDTO(userDTO, hashedPassword))

	return nil
}
