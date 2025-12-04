package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhoubofsy/x-bot/internal/config"
	"google.golang.org/genai"
)

// geminiClient Google Gemini LLM 客户端
type geminiClient struct {
	client *genai.Client
	cfg    *config.LLMConfig
}

// NewGeminiClient 创建 Gemini 客户端
func NewGeminiClient(cfg *config.LLMConfig) Client {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.APIKey,
	})
	if err != nil {
		// 如果初始化失败，返回一个会在调用时报错的客户端
		return &geminiClient{
			client: nil,
			cfg:    cfg,
		}
	}
	return &geminiClient{
		client: client,
		cfg:    cfg,
	}
}

func (c *geminiClient) IsHackathonRelated(ctx context.Context, tweetContent string) (bool, string, error) {
	if c.client == nil {
		return false, "", fmt.Errorf("Gemini client not initialized, check API key")
	}

	prompt := fmt.Sprintf(HackathonDetectionPrompt, tweetContent)

	var lastErr error
	for i := 0; i <= c.cfg.MaxRetries; i++ {
		result, err := c.client.Models.GenerateContent(
			ctx,
			c.cfg.Model,
			genai.Text(prompt),
			nil,
		)
		if err != nil {
			lastErr = err
			continue
		}

		content := result.Text()
		content = strings.TrimSpace(content)
		content = extractJSON(content)

		var detection HackathonDetectionResult
		if err := json.Unmarshal([]byte(content), &detection); err != nil {
			// 如果解析失败，尝试简单判断
			isRelated := strings.Contains(strings.ToLower(content), "true")
			return isRelated, content, nil
		}

		return detection.IsHackathonRelated, content, nil
	}

	return false, "", fmt.Errorf("failed after %d retries: %w", c.cfg.MaxRetries, lastErr)
}
