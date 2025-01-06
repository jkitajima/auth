package user

import "context"

type CreateRequest struct {
	User *User
}

type CreateResponse struct {
	User *User
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (CreateResponse, error) {
	err := s.Repo.Insert(ctx, req.User)
	if err != nil {
		return CreateResponse{nil}, err
	}
	return CreateResponse(req), nil
}
