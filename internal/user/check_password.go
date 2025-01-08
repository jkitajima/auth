package user

import (
	"context"

	"github.com/alexedwards/argon2id"
)

type CheckPasswordRequest struct {
	Input    string
	Password string
}

type CheckPasswordResponse struct {
	Valid bool
}

func (s *Service) CheckPassword(ctx context.Context, req CheckPasswordRequest) (CheckPasswordResponse, error) {
	match, err := argon2id.ComparePasswordAndHash(req.Input, req.Password)
	if err != nil {
		return CheckPasswordResponse{}, err
	}
	return CheckPasswordResponse{Valid: match}, nil
}
