package service

import (
	"context"

	"gpt-team-api/internal/integration/meiguodizhi"
	"gpt-team-api/internal/model"
)

type RandomProfileClient interface {
	FetchProfile(ctx context.Context, cardType model.CardType) (meiguodizhi.ProfileResponse, error)
}

type ProfileService struct {
	client RandomProfileClient
}

func NewProfileService(client RandomProfileClient) *ProfileService {
	return &ProfileService{client: client}
}

func (s *ProfileService) FetchRandomProfile(ctx context.Context) (RandomProfile, error) {
	result, err := s.client.FetchProfile(ctx, model.CardTypeUS)
	if err != nil {
		return RandomProfile{}, err
	}

	return RandomProfile{
		FullName: result.FullName,
		Birthday: result.Birthday,
	}, nil
}
