package auth

import "time"

const (
	refreshTokenDuration = time.Hour * 24 * 14
	accessTokenDuration  = time.Minute * 15
	memory               = 64 * 1024
	iterations           = 3
	threads              = 2
	keyLength            = 32
	hashStructure        = "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	tokenType            = "Bearer"
)
