package user

import "IDIG4110/shared/dto"

func FromCreateDTO(input dto.CreateUserDTO, passwordHash string) User {
	return User{
		Email:        input.Email,
		PasswordHash: passwordHash,
	}
}
