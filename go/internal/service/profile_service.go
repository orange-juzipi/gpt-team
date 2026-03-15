package service

import (
	"context"

	"gpt-team-api/internal/integration/meiguodizhi"
)

type RandomProfileClient interface {
	FetchProfile(ctx context.Context) (meiguodizhi.ProfileResponse, error)
}

type ProfileService struct {
	client RandomProfileClient
}

func NewProfileService(client RandomProfileClient) *ProfileService {
	return &ProfileService{client: client}
}

func (s *ProfileService) FetchRandomProfile(ctx context.Context) (RandomProfile, error) {
	result, err := s.client.FetchProfile(ctx)
	if err != nil {
		return RandomProfile{}, err
	}

	return RandomProfile{
		FullName: result.FullName,
		Birthday: result.Birthday,
	}, nil
}
