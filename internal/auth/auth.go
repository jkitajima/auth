package auth

import (
	"errors"

	"auth/internal/user"
)

var (
	ErrInternal           = errors.New("the auth service encountered an unexpected condition that prevented it from fulfilling the request")
	ErrInvalidCredentials = errors.New("credentials was not valid")
)

type JWTConfig struct {
	Algorithm  string
	Key        string
	Issuer     string
	Audience   []string
	Expiration int
}

type Service struct {
	JWTConfig *JWTConfig
	UserRepo  user.Repoer
}
