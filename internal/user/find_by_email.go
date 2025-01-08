package user

import (
	"context"
)

type FindByEmailRequest struct {
	Email string
}

type FindByEmailResponse struct {
	User *User
}

func (s *Service) FindByEmail(ctx context.Context, req FindByEmailRequest) (FindByEmailResponse, error) {
	user, err := s.Repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return FindByEmailResponse{nil}, err
	}
	return FindByEmailResponse{user}, nil
}
