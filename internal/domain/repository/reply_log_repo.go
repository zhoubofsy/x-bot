package repository

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/domain/entity"
)

type ReplyLogRepository interface {
	// Save 保存回复日志
	Save(ctx context.Context, log *entity.ReplyLog) error

	// GetByTweetID 根据推文ID获取回复日志
	GetByTweetID(ctx context.Context, tweetID string) (*entity.ReplyLog, error)

	// ExistsByTweetID 检查推文是否已处理过
	ExistsByTweetID(ctx context.Context, tweetID string) (bool, error)

	// GetTodayReplyCount 获取今日回复数量
	GetTodayReplyCount(ctx context.Context) (int64, error)

	// GetTodaySuccessCount 获取今日成功回复数量
	GetTodaySuccessCount(ctx context.Context) (int64, error)

	// GetRecentLogs 获取最近的回复日志
	GetRecentLogs(ctx context.Context, limit int) ([]*entity.ReplyLog, error)

	// GetLogsByStatus 根据状态获取回复日志
	GetLogsByStatus(ctx context.Context, status entity.ReplyStatus, limit int) ([]*entity.ReplyLog, error)

	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*ReplyStats, error)
}

type ReplyStats struct {
	TotalCount       int64 `json:"total_count"`
	SuccessCount     int64 `json:"success_count"`
	FailedCount      int64 `json:"failed_count"`
	SkippedCount     int64 `json:"skipped_count"`
	TodayCount       int64 `json:"today_count"`
	TodaySuccessCount int64 `json:"today_success_count"`
	HackathonCount   int64 `json:"hackathon_count"`
}

