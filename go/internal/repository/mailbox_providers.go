package repository

import (
	"context"
	"strings"

	"gpt-team-api/internal/model"

	"gorm.io/gorm"
)

type MailboxProviderRepository struct {
	db *gorm.DB
}

func NewMailboxProviderRepository(db *gorm.DB) *MailboxProviderRepository {
	return &MailboxProviderRepository{db: db}
}

func (r *MailboxProviderRepository) List(ctx context.Context) ([]model.MailboxProvider, error) {
	var items []model.MailboxProvider
	err := r.db.WithContext(ctx).
		Order("domain_suffix ASC").
		Find(&items).Error
	return items, err
}

func (r *MailboxProviderRepository) FindByID(ctx context.Context, id uint64) (model.MailboxProvider, error) {
	var item model.MailboxProvider
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (r *MailboxProviderRepository) FindByDomainSuffix(ctx context.Context, suffix string) (model.MailboxProvider, error) {
	var item model.MailboxProvider
	result := r.db.WithContext(ctx).
		Where("domain_suffix = ?", strings.ToLower(strings.TrimSpace(suffix))).
		Limit(1).
		Find(&item)
	if result.Error != nil {
		return item, result.Error
	}
	if result.RowsAffected == 0 {
		return item, gorm.ErrRecordNotFound
	}
	return item, nil
}

func (r *MailboxProviderRepository) Create(ctx context.Context, item *model.MailboxProvider) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *MailboxProviderRepository) Save(ctx context.Context, item *model.MailboxProvider) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *MailboxProviderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.MailboxProvider{}, id).Error
}
