package auth

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
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

	token, err := s.GenerateToken(ctx, GenerateTokenRequest{UserID: user.ID})
	if err != nil {
		return AccessTokenResponse{}, err
	}

	// privkey, err := jwk.ParseKey(jsonRSAPrivateKey)
	// fmt.Println(err)

	fmt.Println(token.Token)
	signed, err := jwt.Sign(token.Token, jwt.WithKey(jwa.HS256(), []byte("secret")))
	fmt.Println(string(signed))
	fmt.Println(err)

	return AccessTokenResponse{
		AccessToken: string(signed),
	}, nil
}
