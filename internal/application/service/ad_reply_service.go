package service

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
	"github.com/zhoubofsy/x-bot/internal/infrastructure/twitter"
	"go.uber.org/zap"
)

type AdReplyService interface {
	// GetNextAdCopy 获取下一个可用的广告文案
	GetNextAdCopy(ctx context.Context, category string) (*entity.AdCopy, error)

	// ReplyWithAd 在推文下回复广告
	ReplyWithAd(ctx context.Context, tweetID string, adCopy *entity.AdCopy) (*twitter.Tweet, error)
}

type adReplyService struct {
	adCopyRepo    repository.AdCopyRepository
	twitterClient twitter.Client
	logger        *zap.Logger
}

func NewAdReplyService(
	adCopyRepo repository.AdCopyRepository,
	twitterClient twitter.Client,
	logger *zap.Logger,
) AdReplyService {
	return &adReplyService{
		adCopyRepo:    adCopyRepo,
		twitterClient: twitterClient,
		logger:        logger,
	}
}

func (s *adReplyService) GetNextAdCopy(ctx context.Context, category string) (*entity.AdCopy, error) {
	adCopy, err := s.adCopyRepo.GetNextAvailable(ctx, category)
	if err != nil {
		s.logger.Error("获取广告文案失败",
			zap.String("category", category),
			zap.Error(err),
		)
		return nil, err
	}

	return adCopy, nil
}

func (s *adReplyService) ReplyWithAd(ctx context.Context, tweetID string, adCopy *entity.AdCopy) (*twitter.Tweet, error) {
	reply, err := s.twitterClient.ReplyToTweet(ctx, tweetID, adCopy.Content)
	if err != nil {
		s.logger.Error("回复推文失败",
			zap.String("tweet_id", tweetID),
			zap.Error(err),
		)
		return nil, err
	}

	// 增加使用次数
	if err := s.adCopyRepo.IncrementUseCount(ctx, adCopy.ID); err != nil {
		s.logger.Warn("更新广告使用次数失败",
			zap.Int("ad_copy_id", adCopy.ID),
			zap.Error(err),
		)
	}

	s.logger.Info("广告回复成功",
		zap.String("tweet_id", tweetID),
		zap.String("reply_id", reply.ID),
		zap.Int("ad_copy_id", adCopy.ID),
	)

	return reply, nil
}

