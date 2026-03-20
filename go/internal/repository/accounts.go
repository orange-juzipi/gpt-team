package repository

import (
	"context"

	"gpt-team-api/internal/model"

	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) ListTopLevel(ctx context.Context) ([]model.Account, error) {
	var accounts []model.Account
	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Order("created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) ListChildren(ctx context.Context, parentID uint64) ([]model.Account, error) {
	var accounts []model.Account
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) ListChildrenByRelation(ctx context.Context, parentID uint64, relationType model.AccountRelationType) ([]model.Account, error) {
	var accounts []model.Account
	query := r.db.WithContext(ctx).Where("parent_id = ?", parentID)
	if relationType == model.AccountRelationTypeWarranty {
		query = query.Where("(relation_type = ? OR relation_type = '' OR relation_type IS NULL)", relationType)
	} else {
		query = query.Where("relation_type = ?", relationType)
	}

	err := query.
		Order("created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) FindByID(ctx context.Context, id uint64) (model.Account, error) {
	var account model.Account
	err := r.db.WithContext(ctx).First(&account, id).Error
	return account, err
}

func (r *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *AccountRepository) Save(ctx context.Context, account *model.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *AccountRepository) CountChildren(ctx context.Context, parentID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("parent_id = ?", parentID).
		Count(&count).Error
	return count, err
}

func (r *AccountRepository) CountChildrenByRelation(ctx context.Context, parentID uint64, relationType model.AccountRelationType) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("parent_id = ?", parentID)
	if relationType == model.AccountRelationTypeWarranty {
		query = query.Where("(relation_type = ? OR relation_type = '' OR relation_type IS NULL)", relationType)
	} else {
		query = query.Where("relation_type = ?", relationType)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *AccountRepository) DeleteCascade(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("parent_id = ?", id).Delete(&model.Account{}).Error; err != nil {
			return err
		}

		return tx.Delete(&model.Account{}, id).Error
	})
}
