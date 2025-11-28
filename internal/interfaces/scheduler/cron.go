package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/zhoubofsy/x-bot/internal/application/dto"
	"github.com/zhoubofsy/x-bot/internal/application/service"
	"github.com/zhoubofsy/x-bot/internal/config"
	"go.uber.org/zap"
)

type Scheduler struct {
	cron            *cron.Cron
	workflowService service.WorkflowService
	cfg             *config.WorkflowConfig
	logger          *zap.Logger
}

func NewScheduler(
	workflowService service.WorkflowService,
	cfg *config.WorkflowConfig,
	logger *zap.Logger,
) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		workflowService: workflowService,
		cfg:             cfg,
		logger:          logger,
	}
}

func (s *Scheduler) Start() error {
	if !s.cfg.EnableScheduler {
		s.logger.Info("定时任务已禁用")
		return nil
	}

	_, err := s.cron.AddFunc(s.cfg.Schedule, s.executeWorkflow)
	if err != nil {
		s.logger.Error("添加定时任务失败", zap.Error(err))
		return err
	}

	s.cron.Start()
	s.logger.Info("定时任务已启动", zap.String("schedule", s.cfg.Schedule))

	return nil
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		s.logger.Info("定时任务已停止")
	}
}

func (s *Scheduler) executeWorkflow() {
	s.logger.Info("定时任务触发，开始执行工作流")

	ctx := context.Background()
	params := dto.WorkflowParams{
		TweetCount: s.cfg.DefaultTweetCount,
		DryRun:     false,
	}

	result, err := s.workflowService.Execute(ctx, params)
	if err != nil {
		s.logger.Error("定时工作流执行失败", zap.Error(err))
		return
	}

	s.logger.Info("定时工作流执行完成",
		zap.Int("total_users", result.TotalUsers),
		zap.Int("total_tweets", result.TotalTweets),
		zap.Int("hackathon_tweets", result.HackathonTweets),
		zap.Int("successful_replies", result.SuccessfulReplies),
	)
}

