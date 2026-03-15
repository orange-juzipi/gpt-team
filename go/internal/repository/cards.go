package repository

import (
	"context"

	"gpt-team-api/internal/model"

	"gorm.io/gorm"
)

type CardRepository struct {
	db *gorm.DB
}

type CardEventRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) *CardRepository {
	return &CardRepository{db: db}
}

func NewCardEventRepository(db *gorm.DB) *CardEventRepository {
	return &CardEventRepository{db: db}
}

func (r *CardRepository) CreateMany(ctx context.Context, cards []model.Card) error {
	if len(cards) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(&cards).Error
}

func (r *CardRepository) List(ctx context.Context) ([]model.Card, error) {
	var cards []model.Card
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&cards).Error
	return cards, err
}

func (r *CardRepository) FindByID(ctx context.Context, id uint64) (model.Card, error) {
	var card model.Card
	err := r.db.WithContext(ctx).First(&card, id).Error
	return card, err
}

func (r *CardRepository) FindExistingCodes(ctx context.Context, codes []string) ([]string, error) {
	var existing []string
	if len(codes) == 0 {
		return existing, nil
	}

	err := r.db.WithContext(ctx).
		Model(&model.Card{}).
		Where("code IN ?", codes).
		Pluck("code", &existing).Error
	return existing, err
}

func (r *CardRepository) Save(ctx context.Context, card *model.Card) error {
	return r.db.WithContext(ctx).Save(card).Error
}

func (r *CardRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Card{}, id).Error
}

func (r *CardEventRepository) Create(ctx context.Context, event *model.CardEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *CardEventRepository) LatestByCard(ctx context.Context, cardID uint64) (map[model.CardEventType]model.CardEvent, error) {
	var events []model.CardEvent
	err := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Order("created_at DESC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	result := make(map[model.CardEventType]model.CardEvent, len(events))
	for _, event := range events {
		if _, exists := result[event.EventType]; exists {
			continue
		}

		result[event.EventType] = event
	}

	return result, nil
}
