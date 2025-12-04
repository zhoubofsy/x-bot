package service

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/application/dto"
	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
	"github.com/zhoubofsy/x-bot/internal/infrastructure/twitter"
	"go.uber.org/zap"
)

type FollowerService interface {
	// SyncFollowing 同步当前用户的关注列表到数据库
	SyncFollowing(ctx context.Context) (*dto.SyncFollowingResult, error)

	// GetAllActiveFollowers 获取所有活跃的关注用户
	GetAllActiveFollowers(ctx context.Context) ([]*entity.FollowedUser, error)
}

type followerService struct {
	userRepo      repository.UserRepository
	twitterClient twitter.Client
	logger        *zap.Logger
}

func NewFollowerService(
	userRepo repository.UserRepository,
	twitterClient twitter.Client,
	logger *zap.Logger,
) FollowerService {
	return &followerService{
		userRepo:      userRepo,
		twitterClient: twitterClient,
		logger:        logger,
	}
}

func (s *followerService) SyncFollowing(ctx context.Context) (*dto.SyncFollowingResult, error) {
	result := &dto.SyncFollowingResult{}

	// 获取当前用户信息
	me, err := s.twitterClient.GetMe(ctx)
	if err != nil {
		s.logger.Error("获取当前用户信息失败", zap.Error(err))
		// 如果获取用户信息失败，尝试返回数据库中已有的用户
		return s.fallbackToDatabase(ctx, result, err)
	}

	// 获取关注列表
	following, err := s.twitterClient.GetFollowing(ctx, me.ID)
	if err != nil {
		s.logger.Warn("获取关注列表失败，回退到数据库中的手动添加用户", zap.Error(err))
		// 如果 API 失败（如 Free 套餐限制），返回数据库中已有的用户
		return s.fallbackToDatabase(ctx, result, err)
	}

	result.TotalCount = len(following)

	// 转换并保存到数据库
	users := make([]*entity.FollowedUser, 0, len(following))
	for _, f := range following {
		users = append(users, &entity.FollowedUser{
			TwitterUserID: f.ID,
			Username:      f.Username,
			DisplayName:   f.Name,
			IsActive:      true,
		})
	}

	if err := s.userRepo.BatchSave(ctx, users); err != nil {
		s.logger.Error("保存关注用户失败", zap.Error(err))
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	result.Source = "twitter_api"
	s.logger.Info("同步关注列表完成",
		zap.Int("total", result.TotalCount),
	)

	return result, nil
}

// fallbackToDatabase 当 Twitter API 失败时，回退到数据库中已有的用户
func (s *followerService) fallbackToDatabase(ctx context.Context, result *dto.SyncFollowingResult, apiErr error) (*dto.SyncFollowingResult, error) {
	users, err := s.userRepo.GetAllActiveUsers(ctx)
	if err != nil {
		s.logger.Error("获取数据库用户失败", zap.Error(err))
		return nil, apiErr // 返回原始 API 错误
	}

	if len(users) == 0 {
		result.Errors = append(result.Errors, "Twitter API 不可用且数据库中无手动添加的用户，请先通过 /api/v1/users 接口添加要监控的用户")
		return result, apiErr
	}

	result.TotalCount = len(users)
	result.Source = "database"
	result.Errors = append(result.Errors, "Twitter API 不可用，已使用数据库中的手动添加用户: "+apiErr.Error())

	s.logger.Info("使用数据库中的手动添加用户",
		zap.Int("total", result.TotalCount),
	)

	return result, nil // 返回 nil error，表示可以继续工作流
}

func (s *followerService) GetAllActiveFollowers(ctx context.Context) ([]*entity.FollowedUser, error) {
	return s.userRepo.GetAllActiveUsers(ctx)
}
