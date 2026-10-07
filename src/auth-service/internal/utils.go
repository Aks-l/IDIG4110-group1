package internal

import "crypto/rand"

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)

	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	return salt, nil
}
