package postgres

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAllActiveUsers(ctx context.Context) ([]*entity.FollowedUser, error) {
	var users []*entity.FollowedUser
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&users).Error
	return users, err
}

func (r *userRepository) GetByTwitterID(ctx context.Context, twitterID string) (*entity.FollowedUser, error) {
	var user entity.FollowedUser
	err := r.db.WithContext(ctx).Where("twitter_user_id = ?", twitterID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Save(ctx context.Context, user *entity.FollowedUser) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "twitter_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "display_name", "is_active", "updated_at"}),
	}).Create(user).Error
}

func (r *userRepository) BatchSave(ctx context.Context, users []*entity.FollowedUser) error {
	if len(users) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "twitter_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "display_name", "is_active", "updated_at"}),
	}).CreateInBatches(users, 100).Error
}

func (r *userRepository) UpdateActiveStatus(ctx context.Context, twitterID string, isActive bool) error {
	return r.db.WithContext(ctx).Model(&entity.FollowedUser{}).
		Where("twitter_user_id = ?", twitterID).
		Update("is_active", isActive).Error
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.FollowedUser{}, id).Error
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.FollowedUser{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

