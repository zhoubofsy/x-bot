package service

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/infrastructure/twitter"
	"go.uber.org/zap"
)

type TweetService interface {
	// GetUserTweets 获取指定用户的最新推文
	GetUserTweets(ctx context.Context, userID string, count int) ([]twitter.Tweet, error)
}

type tweetService struct {
	twitterClient twitter.Client
	logger        *zap.Logger
}

func NewTweetService(
	twitterClient twitter.Client,
	logger *zap.Logger,
) TweetService {
	return &tweetService{
		twitterClient: twitterClient,
		logger:        logger,
	}
}

func (s *tweetService) GetUserTweets(ctx context.Context, userID string, count int) ([]twitter.Tweet, error) {
	tweets, err := s.twitterClient.GetUserTweets(ctx, userID, count)
	if err != nil {
		s.logger.Error("获取用户推文失败",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, err
	}

	s.logger.Debug("获取用户推文成功",
		zap.String("user_id", userID),
		zap.Int("count", len(tweets)),
	)

	return tweets, nil
}

