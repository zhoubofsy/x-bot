package service

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/application/dto"
	"github.com/zhoubofsy/x-bot/internal/config"
	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
	"github.com/zhoubofsy/x-bot/internal/infrastructure/twitter"
	"go.uber.org/zap"
)

type WorkflowService interface {
	// Execute 执行完整工作流
	Execute(ctx context.Context, params dto.WorkflowParams) (*dto.WorkflowResult, error)
}

type workflowService struct {
	followerService   FollowerService
	tweetService      TweetService
	hackathonDetector HackathonDetector
	adReplyService    AdReplyService
	replyLogRepo      repository.ReplyLogRepository
	cfg               *config.WorkflowConfig
	logger            *zap.Logger
}

func NewWorkflowService(
	followerService FollowerService,
	tweetService TweetService,
	hackathonDetector HackathonDetector,
	adReplyService AdReplyService,
	replyLogRepo repository.ReplyLogRepository,
	cfg *config.WorkflowConfig,
	logger *zap.Logger,
) WorkflowService {
	return &workflowService{
		followerService:   followerService,
		tweetService:      tweetService,
		hackathonDetector: hackathonDetector,
		adReplyService:    adReplyService,
		replyLogRepo:      replyLogRepo,
		cfg:               cfg,
		logger:            logger,
	}
}

func (s *workflowService) Execute(ctx context.Context, params dto.WorkflowParams) (*dto.WorkflowResult, error) {
	result := &dto.WorkflowResult{}

	// 使用默认值
	if params.TweetCount <= 0 {
		params.TweetCount = s.cfg.DefaultTweetCount
	}

	s.logger.Info("开始执行工作流",
		zap.Int("tweet_count", params.TweetCount),
		zap.Bool("dry_run", params.DryRun),
	)

	// Step 1: 获取所有活跃关注用户
	users, err := s.followerService.GetAllActiveFollowers(ctx)
	if err != nil {
		s.logger.Error("获取关注用户失败", zap.Error(err))
		return nil, err
	}
	result.TotalUsers = len(users)

	s.logger.Info("获取关注用户完成", zap.Int("count", len(users)))

	// Step 2 & 3 & 4: 遍历用户并处理推文
	for _, user := range users {
		if err := s.processUserTweets(ctx, user, params, result); err != nil {
			s.logger.Error("处理用户推文失败",
				zap.String("user_id", user.TwitterUserID),
				zap.Error(err),
			)
			result.Errors = append(result.Errors, err.Error())
		}
	}

	s.logger.Info("工作流执行完成",
		zap.Int("total_users", result.TotalUsers),
		zap.Int("total_tweets", result.TotalTweets),
		zap.Int("hackathon_tweets", result.HackathonTweets),
		zap.Int("successful_replies", result.SuccessfulReplies),
		zap.Int("failed_replies", result.FailedReplies),
		zap.Int("skipped_tweets", result.SkippedTweets),
	)

	return result, nil
}

func (s *workflowService) processUserTweets(
	ctx context.Context,
	user *entity.FollowedUser,
	params dto.WorkflowParams,
	result *dto.WorkflowResult,
) error {
	// 获取用户推文
	tweets, err := s.tweetService.GetUserTweets(ctx, user.TwitterUserID, params.TweetCount)
	if err != nil {
		return err
	}
	result.TotalTweets += len(tweets)

	// 处理每条推文
	for _, tweet := range tweets {
		processResult := s.processSingleTweet(ctx, tweet, params.DryRun)
		s.updateResult(result, processResult)

		// 回复间隔
		// if processResult.Success && !params.DryRun {
		// 	time.Sleep(s.cfg.ReplyInterval)
		// }
	}

	return nil
}

func (s *workflowService) processSingleTweet(
	ctx context.Context,
	tweet twitter.Tweet,
	dryRun bool,
) dto.ProcessResult {
	pr := dto.ProcessResult{TweetID: tweet.ID}

	// 检查是否已处理过
	exists, _ := s.replyLogRepo.ExistsByTweetID(ctx, tweet.ID)
	if exists {
		pr.Skipped = true
		return pr
	}

	// 检查今日回复限制
	todayCount, _ := s.replyLogRepo.GetTodaySuccessCount(ctx)
	if todayCount >= int64(s.cfg.MaxDailyReplies) {
		s.logger.Warn("已达到今日回复上限")
		pr.Skipped = true
		return pr
	}

	// LLM 检测
	isHackathon, llmResponse, err := s.hackathonDetector.Detect(ctx, tweet.Text)
	if err != nil {
		pr.Error = err
		// s.saveReplyLog(ctx, tweet, "", nil, entity.ReplyStatusFailed, llmResponse, false, err.Error())
		return pr
	}

	pr.IsHackathon = isHackathon

	if !isHackathon {
		pr.Skipped = true
		// s.saveReplyLog(ctx, tweet, "", nil, entity.ReplyStatusSkipped, llmResponse, false, "")
		return pr
	}

	if dryRun {
		s.saveReplyLog(ctx, tweet, "", nil, entity.ReplyStatusDryRun, llmResponse, true, "")
		return pr
	}

	// 获取广告并回复
	adCopy, err := s.adReplyService.GetNextAdCopy(ctx, "hackathon")
	if err != nil {
		pr.Error = err
		//s.saveReplyLog(ctx, tweet, "", nil, entity.ReplyStatusFailed, llmResponse, true, err.Error())
		return pr
	}

	replyTweet, err := s.adReplyService.ReplyWithAd(ctx, tweet.ID, adCopy)
	if err != nil {
		pr.Error = err
		//s.saveReplyLog(ctx, tweet, "", &adCopy.ID, entity.ReplyStatusFailed, llmResponse, true, err.Error())
		return pr
	}

	pr.Success = true
	s.saveReplyLog(ctx, tweet, replyTweet.ID, &adCopy.ID, entity.ReplyStatusSuccess, llmResponse, true, "")

	return pr
}

func (s *workflowService) saveReplyLog(
	ctx context.Context,
	tweet twitter.Tweet,
	replyTweetID string,
	adCopyID *int,
	status entity.ReplyStatus,
	llmResponse string,
	isHackathon bool,
	errorMsg string,
) {
	log := &entity.ReplyLog{
		TweetID:       tweet.ID,
		TweetAuthorID: tweet.AuthorID,
		TweetContent:  tweet.Text,
		ReplyTweetID:  replyTweetID,
		AdCopyID:      adCopyID,
		Status:        status,
		LLMResponse:   llmResponse,
		IsHackathon:   isHackathon,
		ErrorMessage:  errorMsg,
	}

	if err := s.replyLogRepo.Save(ctx, log); err != nil {
		s.logger.Error("保存回复日志失败",
			zap.String("tweet_id", tweet.ID),
			zap.Error(err),
		)
	}
}

func (s *workflowService) updateResult(result *dto.WorkflowResult, pr dto.ProcessResult) {
	if pr.IsHackathon {
		result.HackathonTweets++
	}
	if pr.Success {
		result.SuccessfulReplies++
	}
	if pr.Skipped {
		result.SkippedTweets++
	}
	if pr.Error != nil {
		result.FailedReplies++
		result.Errors = append(result.Errors, pr.Error.Error())
	}
}
