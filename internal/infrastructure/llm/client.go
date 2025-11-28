package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/zhoubofsy/x-bot/internal/config"
)

type Client interface {
	// IsHackathonRelated 判断推文是否与黑客松相关
	// 返回：是否相关、原始LLM响应、错误
	IsHackathonRelated(ctx context.Context, tweetContent string) (bool, string, error)
}

type client struct {
	httpClient *http.Client
	cfg        *config.LLMConfig
}

func NewClient(cfg *config.LLMConfig) Client {
	return &client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		cfg:        cfg,
	}
}

func (c *client) IsHackathonRelated(ctx context.Context, tweetContent string) (bool, string, error) {
	prompt := fmt.Sprintf(HackathonDetectionPrompt, tweetContent)

	req := ChatRequest{
		Model: c.cfg.Model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.1,
		MaxTokens:   500,
	}

	respBody, err := c.doRequest(ctx, req)
	if err != nil {
		return false, "", err
	}

	// 解析LLM响应
	if len(respBody.Choices) == 0 {
		return false, "", fmt.Errorf("empty response from LLM")
	}

	content := respBody.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// 尝试提取JSON部分
	content = extractJSON(content)

	var result HackathonDetectionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// 如果解析失败，尝试简单判断
		isRelated := strings.Contains(strings.ToLower(content), "true")
		return isRelated, content, nil
	}

	return result.IsHackathonRelated, content, nil
}

func (c *client) doRequest(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.cfg.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	var lastErr error
	for i := 0; i <= c.cfg.MaxRetries; i++ {
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("LLM API error: status=%d, body=%s", resp.StatusCode, string(respBody))
			continue
		}

		var chatResp ChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		return &chatResp, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.cfg.MaxRetries, lastErr)
}

// extractJSON 从文本中提取JSON部分
func extractJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 && end > start {
		return text[start : end+1]
	}
	return text
}

