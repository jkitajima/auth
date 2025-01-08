package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type GenerateTokenRequest struct {
	UserID uuid.UUID
}

type GenerateTokenResponse struct {
	AccessToken []byte
	TokenType   string
	ExpiresIn   int
}

func (s *Service) GenerateToken(ctx context.Context, req GenerateTokenRequest) (GenerateTokenResponse, error) {
	now := time.Now()

	token, err := jwt.NewBuilder().
		Issuer(s.JWTConfig.Issuer).
		Subject(req.UserID.String()).
		Audience(s.JWTConfig.Audience).
		Expiration(now.Add(time.Duration(s.JWTConfig.Expiration) * time.Second)).
		NotBefore(now).
		IssuedAt(now).
		JwtID(uuid.NewString()).
		Build()
	if err != nil {
		return GenerateTokenResponse{}, err
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS256(), []byte(s.JWTConfig.Key)))
	if err != nil {
		return GenerateTokenResponse{}, err
	}

	return GenerateTokenResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   s.JWTConfig.Expiration,
	}, nil
}
