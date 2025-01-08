package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type GenerateTokenRequest struct {
	UserID uuid.UUID
}

type GenerateTokenResponse struct {
	Token jwt.Token
}

func (s *Service) GenerateToken(ctx context.Context, req GenerateTokenRequest) (GenerateTokenResponse, error) {
	now := time.Now()

	token, err := jwt.NewBuilder().
		Issuer(s.JWTConfig.Issuer).
		Subject(req.UserID.String()).
		Audience(s.JWTConfig.Audience).
		Expiration(now.Add(time.Duration(s.JWTConfig.Expiration) * time.Second)).
		IssuedAt(now).
		JwtID(uuid.NewString()).
		Build()
	if err != nil {
		return GenerateTokenResponse{}, err
	}

	return GenerateTokenResponse{Token: token}, nil
}
