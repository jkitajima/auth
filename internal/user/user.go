package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInternal           = errors.New("the user service encountered an unexpected condition that prevented it from fulfilling the request")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFoundByID       = errors.New("could not find any user with provided ID")
	ErrNotFoundByEmail    = errors.New("could not find any user with provided email")
	ErrEmailAlreadyInUse  = errors.New("provided email address is already in use")
)

type User struct {
	ID                         uuid.UUID
	Email                      string
	EmailVerified              bool
	Password                   string
	VerificationCode           *string
	VerificationCodeExpiration *time.Time
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	DeletedAt                  *time.Time
}

type Service struct {
	Repo Repoer
}

type Repoer interface {
	Insert(context.Context, *User) error
	FindByID(context.Context, uuid.UUID) (*User, error)
	FindByEmail(context.Context, string) (*User, error)
	HardDeleteByID(context.Context, uuid.UUID) error
}
