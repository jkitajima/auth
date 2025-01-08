package auth

import (
	"context"
)

type AccessTokenRequest struct {
	Username string
	Password string
}

type AccessTokenResponse struct {
	AccessToken string
}

func (s *Service) RequestAccessToken(ctx context.Context, req AccessTokenRequest) (AccessTokenResponse, error) {
	// Find user by username (email)
	user, err := s.UserRepo.FindByEmail(ctx, req.Username)
	if err != nil {
		return AccessTokenResponse{}, err
	}

	// Check if incoming password matches stored password
	checkPasswordRequest := CheckPasswordRequest{
		Input:    req.Password,
		Password: user.Password,
	}
	checkPasswordResponse, err := s.CheckPassword(ctx, checkPasswordRequest)
	if err != nil {
		return AccessTokenResponse{}, err
	}

	// If match is not valid, then deny access token request
	if !checkPasswordResponse.Valid {
		return AccessTokenResponse{}, ErrInvalidCredentials
	}

	return AccessTokenResponse{
		AccessToken: "test_access_token",
	}, nil
}
