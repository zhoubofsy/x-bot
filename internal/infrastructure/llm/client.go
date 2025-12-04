package llm

import (
	"context"
	"strings"

	"github.com/zhoubofsy/x-bot/internal/config"
)

// Client LLM 客户端接口
type Client interface {
	// IsHackathonRelated 判断推文是否与黑客松相关
	// 返回：是否相关、原始LLM响应、错误
	IsHackathonRelated(ctx context.Context, tweetContent string) (bool, string, error)
}

// NewClient 根据配置创建对应的 LLM 客户端
func NewClient(cfg *config.LLMConfig) Client {
	switch strings.ToLower(cfg.Provider) {
	case "gemini", "google":
		return NewGeminiClient(cfg)
	case "openai", "":
		return NewOpenAIClient(cfg)
	default:
		// 默认使用 OpenAI 兼容接口（也适用于 Groq、Together AI 等）
		return NewOpenAIClient(cfg)
	}
}
