package user

import (
	"context"

	"github.com/google/uuid"
)

type FindByIDRequest struct {
	ID uuid.UUID
}

type FindByIDResponse struct {
	User *User
}

func (s *Service) FindByID(ctx context.Context, req FindByIDRequest) (FindByIDResponse, error) {
	user, err := s.Repo.FindByID(ctx, req.ID)
	if err != nil {
		return FindByIDResponse{nil}, err
	}
	return FindByIDResponse{user}, nil
}
