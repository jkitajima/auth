package user

import (
	"context"

	"auth/pkg/password"

	"github.com/google/uuid"
)

type HardDeleteByIDRequest struct {
	ID       uuid.UUID
	Password string
}

func (s *Service) HardDeleteByID(ctx context.Context, req HardDeleteByIDRequest) error {
	// Check if the user exists first
	findResponse, err := s.FindByID(ctx, FindByIDRequest{req.ID})
	if err != nil {
		return err
	}

	// Check if incoming password matches stored password
	checkPasswordRequest := password.CheckPasswordRequest{
		Input:    req.Password,
		Password: findResponse.User.Password,
	}
	checkPasswordResponse, err := password.CheckPassword(ctx, checkPasswordRequest)
	if err != nil {
		return err
	}

	// If match is not valid, then deny access token request
	if !checkPasswordResponse.Valid {
		return ErrInvalidCredentials
	}

	err = s.Repo.HardDeleteByID(ctx, req.ID)
	if err != nil {
		return err
	}
	return nil
}
