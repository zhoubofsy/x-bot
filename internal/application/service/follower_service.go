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
		return nil, err
	}

	// 获取关注列表
	following, err := s.twitterClient.GetFollowing(ctx, me.ID)
	if err != nil {
		s.logger.Error("获取关注列表失败", zap.Error(err))
		return nil, err
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

	s.logger.Info("同步关注列表完成",
		zap.Int("total", result.TotalCount),
	)

	return result, nil
}

func (s *followerService) GetAllActiveFollowers(ctx context.Context) ([]*entity.FollowedUser, error) {
	return s.userRepo.GetAllActiveUsers(ctx)
}

