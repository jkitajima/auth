package auth

import (
	"errors"

	"auth/internal/user"
)

var (
	ErrInternal           = errors.New("the auth service encountered an unexpected condition that prevented it from fulfilling the request")
	ErrInvalidCredentials = errors.New("credentials was not valid")
)

type Service struct {
	UserRepo user.Repoer
}
