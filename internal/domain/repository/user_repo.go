package repository

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/domain/entity"
)

type UserRepository interface {
	// GetAllActiveUsers 获取所有活跃的关注用户
	GetAllActiveUsers(ctx context.Context) ([]*entity.FollowedUser, error)

	// GetByTwitterID 根据Twitter用户ID获取用户
	GetByTwitterID(ctx context.Context, twitterID string) (*entity.FollowedUser, error)

	// Save 保存用户
	Save(ctx context.Context, user *entity.FollowedUser) error

	// BatchSave 批量保存用户
	BatchSave(ctx context.Context, users []*entity.FollowedUser) error

	// UpdateActiveStatus 更新用户活跃状态
	UpdateActiveStatus(ctx context.Context, twitterID string, isActive bool) error

	// Delete 删除用户
	Delete(ctx context.Context, id int) error

	// Count 获取用户总数
	Count(ctx context.Context) (int64, error)
}

