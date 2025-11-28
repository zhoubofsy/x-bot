package postgres

import (
	"context"
	"time"

	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
	"gorm.io/gorm"
)

type adCopyRepository struct {
	db *gorm.DB
}

func NewAdCopyRepository(db *gorm.DB) repository.AdCopyRepository {
	return &adCopyRepository{db: db}
}

func (r *adCopyRepository) GetAll(ctx context.Context) ([]*entity.AdCopy, error) {
	var adCopies []*entity.AdCopy
	err := r.db.WithContext(ctx).Order("priority DESC, created_at DESC").Find(&adCopies).Error
	return adCopies, err
}

func (r *adCopyRepository) GetByID(ctx context.Context, id int) (*entity.AdCopy, error) {
	var adCopy entity.AdCopy
	err := r.db.WithContext(ctx).First(&adCopy, id).Error
	if err != nil {
		return nil, err
	}
	return &adCopy, nil
}

func (r *adCopyRepository) GetActiveByCategory(ctx context.Context, category string) ([]*entity.AdCopy, error) {
	var adCopies []*entity.AdCopy
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND category = ?", true, category).
		Order("priority DESC, use_count ASC").
		Find(&adCopies).Error
	return adCopies, err
}

func (r *adCopyRepository) GetNextAvailable(ctx context.Context, category string) (*entity.AdCopy, error) {
	var adCopy entity.AdCopy
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND category = ?", true, category).
		Order("priority DESC, use_count ASC, RANDOM()").
		First(&adCopy).Error
	if err != nil {
		return nil, err
	}
	return &adCopy, nil
}

func (r *adCopyRepository) IncrementUseCount(ctx context.Context, id int) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.AdCopy{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"use_count":    gorm.Expr("use_count + 1"),
			"last_used_at": now,
			"updated_at":   now,
		}).Error
}

func (r *adCopyRepository) Save(ctx context.Context, adCopy *entity.AdCopy) error {
	return r.db.WithContext(ctx).Create(adCopy).Error
}

func (r *adCopyRepository) Update(ctx context.Context, adCopy *entity.AdCopy) error {
	return r.db.WithContext(ctx).Save(adCopy).Error
}

func (r *adCopyRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.AdCopy{}, id).Error
}

func (r *adCopyRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.AdCopy{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

