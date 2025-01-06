package auth

import (
	"errors"

	"auth/internal/user"
)

type Service struct {
	UserRepo user.Repoer
}

var ErrInternal = errors.New("the auth service encountered an unexpected condition that prevented it from fulfilling the request")
