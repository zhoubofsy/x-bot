package postgres

import (
	"context"
	"time"

	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
	"gorm.io/gorm"
)

type replyLogRepository struct {
	db *gorm.DB
}

func NewReplyLogRepository(db *gorm.DB) repository.ReplyLogRepository {
	return &replyLogRepository{db: db}
}

func (r *replyLogRepository) Save(ctx context.Context, log *entity.ReplyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *replyLogRepository) GetByTweetID(ctx context.Context, tweetID string) (*entity.ReplyLog, error) {
	var log entity.ReplyLog
	err := r.db.WithContext(ctx).Where("tweet_id = ?", tweetID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *replyLogRepository) ExistsByTweetID(ctx context.Context, tweetID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Where("tweet_id = ?", tweetID).Count(&count).Error
	return count > 0, err
}

func (r *replyLogRepository) GetTodayReplyCount(ctx context.Context) (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Model(&entity.ReplyLog{}).
		Where("created_at >= ?", today).
		Count(&count).Error
	return count, err
}

func (r *replyLogRepository) GetTodaySuccessCount(ctx context.Context) (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Model(&entity.ReplyLog{}).
		Where("created_at >= ? AND status = ?", today, entity.ReplyStatusSuccess).
		Count(&count).Error
	return count, err
}

func (r *replyLogRepository) GetRecentLogs(ctx context.Context, limit int) ([]*entity.ReplyLog, error) {
	var logs []*entity.ReplyLog
	err := r.db.WithContext(ctx).
		Preload("AdCopy").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *replyLogRepository) GetLogsByStatus(ctx context.Context, status entity.ReplyStatus, limit int) ([]*entity.ReplyLog, error) {
	var logs []*entity.ReplyLog
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *replyLogRepository) GetStats(ctx context.Context) (*repository.ReplyStats, error) {
	stats := &repository.ReplyStats{}
	today := time.Now().Truncate(24 * time.Hour)

	// Total count
	r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Count(&stats.TotalCount)

	// Success count
	r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Where("status = ?", entity.ReplyStatusSuccess).Count(&stats.SuccessCount)

	// Failed count
	r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Where("status = ?", entity.ReplyStatusFailed).Count(&stats.FailedCount)

	// Skipped count
	r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Where("status = ?", entity.ReplyStatusSkipped).Count(&stats.SkippedCount)

	// Today count
	r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Where("created_at >= ?", today).Count(&stats.TodayCount)

	// Today success count
	r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Where("created_at >= ? AND status = ?", today, entity.ReplyStatusSuccess).Count(&stats.TodaySuccessCount)

	// Hackathon count
	r.db.WithContext(ctx).Model(&entity.ReplyLog{}).Where("is_hackathon = ?", true).Count(&stats.HackathonCount)

	return stats, nil
}

