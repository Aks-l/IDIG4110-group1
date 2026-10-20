package auth

import "IDIG4110/shared/dto"

func FromCreateDTO(input dto.UserDTO, passwordHash string) User {
	return User{
		Email:        input.Email,
		PasswordHash: passwordHash,
	}
}
