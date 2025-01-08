package auth

import (
	"context"

	"auth/pkg/password"
)

type AccessTokenRequest struct {
	Username string
	Password string
}

type AccessTokenResponse struct {
	GenerateTokenResponse
}

func (s *Service) RequestAccessToken(ctx context.Context, req AccessTokenRequest) (AccessTokenResponse, error) {
	// Find user by username (email)
	user, err := s.UserRepo.FindByEmail(ctx, req.Username)
	if err != nil {
		return AccessTokenResponse{}, err
	}

	// Check if incoming password matches stored password
	checkPasswordRequest := password.CheckPasswordRequest{
		Input:    req.Password,
		Password: user.Password,
	}
	checkPasswordResponse, err := password.CheckPassword(ctx, checkPasswordRequest)
	if err != nil {
		return AccessTokenResponse{}, err
	}

	// If match is not valid, then deny access token request
	if !checkPasswordResponse.Valid {
		return AccessTokenResponse{}, ErrInvalidCredentials
	}

	token, err := s.GenerateToken(ctx, GenerateTokenRequest{UserID: user.ID})
	if err != nil {
		return AccessTokenResponse{}, err
	}

	return AccessTokenResponse{token}, nil
}
