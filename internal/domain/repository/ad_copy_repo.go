package repository

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/domain/entity"
)

type AdCopyRepository interface {
	// GetAll 获取所有广告文案
	GetAll(ctx context.Context) ([]*entity.AdCopy, error)

	// GetByID 根据ID获取广告文案
	GetByID(ctx context.Context, id int) (*entity.AdCopy, error)

	// GetActiveByCategory 获取指定类别的活跃广告文案
	GetActiveByCategory(ctx context.Context, category string) ([]*entity.AdCopy, error)

	// GetNextAvailable 获取下一个可用的广告文案（按优先级和使用次数）
	GetNextAvailable(ctx context.Context, category string) (*entity.AdCopy, error)

	// IncrementUseCount 增加使用次数
	IncrementUseCount(ctx context.Context, id int) error

	// Save 保存广告文案
	Save(ctx context.Context, adCopy *entity.AdCopy) error

	// Update 更新广告文案
	Update(ctx context.Context, adCopy *entity.AdCopy) error

	// Delete 删除广告文案
	Delete(ctx context.Context, id int) error

	// Count 获取广告文案总数
	Count(ctx context.Context) (int64, error)
}

