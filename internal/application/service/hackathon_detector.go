package service

import (
	"context"

	"github.com/zhoubofsy/x-bot/internal/infrastructure/llm"
	"go.uber.org/zap"
)

type HackathonDetector interface {
	// Detect 检测推文是否与黑客松相关
	// 返回：是否相关、LLM原始响应、错误
	Detect(ctx context.Context, tweetContent string) (bool, string, error)
}

type hackathonDetector struct {
	llmClient llm.Client
	logger    *zap.Logger
}

func NewHackathonDetector(
	llmClient llm.Client,
	logger *zap.Logger,
) HackathonDetector {
	return &hackathonDetector{
		llmClient: llmClient,
		logger:    logger,
	}
}

func (d *hackathonDetector) Detect(ctx context.Context, tweetContent string) (bool, string, error) {
	isRelated, rawResponse, err := d.llmClient.IsHackathonRelated(ctx, tweetContent)
	if err != nil {
		d.logger.Error("LLM检测失败", zap.Error(err))
		return false, "", err
	}

	d.logger.Debug("黑客松检测结果",
		zap.Bool("is_related", isRelated),
		zap.String("response", rawResponse),
	)

	return isRelated, rawResponse, nil
}

