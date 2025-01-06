package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
	// FindByID(context.Context, uuid.UUID) (*User, error)
	// UpdateByID(context.Context, uuid.UUID, *User) error
	// SoftDeleteByID(context.Context, uuid.UUID) error
}

var (
	ErrInternal          = errors.New("the user service encountered an unexpected condition that prevented it from fulfilling the request")
	ErrNotFoundByID      = errors.New("could not find any user with provided ID")
	ErrEmailAlreadyInUse = errors.New("provided email address is already in use")
)
