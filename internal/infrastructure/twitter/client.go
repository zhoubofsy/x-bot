package twitter

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zhoubofsy/x-bot/internal/config"
)

const (
	baseURL  = "https://api.x.com/2"
	oauthURL = "https://api.x.com"
)

type Client interface {
	GetMe(ctx context.Context) (*TwitterUser, error)
	GetFollowing(ctx context.Context, userID string) ([]TwitterUser, error)
	GetUserTweets(ctx context.Context, userID string, maxResults int) ([]Tweet, error)
	ReplyToTweet(ctx context.Context, tweetID string, text string) (*Tweet, error)
}

type client struct {
	httpClient  *http.Client
	cfg         *config.TwitterConfig
	currentUser *TwitterUser
}

func NewClient(cfg *config.TwitterConfig) Client {
	return &client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		cfg:        cfg,
	}
}

func (c *client) GetMe(ctx context.Context) (*TwitterUser, error) {
	if c.currentUser != nil {
		return c.currentUser, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/users/me", nil)
	if err != nil {
		return nil, err
	}

	c.signRequest(req, "GET", baseURL+"/users/me", nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var result struct {
		Data TwitterUser `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	c.currentUser = &result.Data
	return c.currentUser, nil
}

func (c *client) GetFollowing(ctx context.Context, userID string) ([]TwitterUser, error) {
	var allUsers []TwitterUser
	nextToken := ""

	for {
		endpoint := fmt.Sprintf("%s/users/%s/following?max_results=100", baseURL, userID)
		if nextToken != "" {
			endpoint += "&pagination_token=" + nextToken
		}

		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		c.signRequest(req, "GET", endpoint, nil)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, c.handleError(resp)
		}

		var result FollowingResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		allUsers = append(allUsers, result.Data...)

		if result.Meta == nil || result.Meta.NextToken == "" {
			break
		}
		nextToken = result.Meta.NextToken
	}

	return allUsers, nil
}

func (c *client) GetUserTweets(ctx context.Context, userID string, maxResults int) ([]Tweet, error) {
	if maxResults > 100 {
		maxResults = 100
	}
	if maxResults < 5 {
		maxResults = 5
	}

	endpoint := fmt.Sprintf("%s/users/%s/tweets?max_results=%d&tweet.fields=created_at,author_id",
		baseURL, userID, maxResults)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	c.signRequest(req, "GET", endpoint, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
		// fake response for testing
		// fakeData := `{
		// 	"data": [
		// 		{
		// 			"id": "1996142611418304942",
		// 			"text": "预测市场专项黑客松正在进行中，12月12日截止\n\n获奖项目将通过https://t.co/l9bADiBPe9获得资金支持、导师指导，以及市场推广支持，并在评委、合作伙伴和 https://t.co/l9bADiBPe9 社区面前做直播Demo",
		// 			"created_at": "2025-12-03T10:00:00Z",
		// 			"author_id": "4239722354"
		// 		}
		// 	]
		// }`
		// resp.Body = io.NopCloser(strings.NewReader(fakeData))
	}

	var result TweetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *client) ReplyToTweet(ctx context.Context, tweetID string, text string) (*Tweet, error) {
	endpoint := baseURL + "/tweets"

	payload := CreateTweetRequest{
		Text: text,
		Reply: &ReplyConfig{
			InReplyToTweetID: tweetID,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 对于 POST 请求，OAuth 签名不包含 body 内容
	// 参考: https://github.com/xdevplatform/samples/blob/main/python/posts/create_tweet.py
	c.signRequestForPost(req, endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleError(resp)
	}

	var result CreateTweetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Tweet{
		ID:   result.Data.ID,
		Text: result.Data.Text,
	}, nil
}

// signRequestForPost 专门用于 POST 请求的 OAuth 签名
// POST 请求的 JSON body 不参与签名计算
func (c *client) signRequestForPost(req *http.Request, endpoint string) {
	oauthParams := map[string]string{
		"oauth_consumer_key":     c.cfg.APIKey,
		"oauth_nonce":            uuid.New().String(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_token":            c.cfg.AccessToken,
		"oauth_version":          "1.0",
	}

	// 对于 POST JSON 请求，只使用 oauth 参数进行签名
	var keys []string
	for k := range oauthParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, url.QueryEscape(k)+"="+url.QueryEscape(oauthParams[k]))
	}
	paramString := strings.Join(paramPairs, "&")

	// 签名基础字符串
	signatureBase := "POST&" + url.QueryEscape(endpoint) + "&" + url.QueryEscape(paramString)

	// 签名密钥
	signingKey := url.QueryEscape(c.cfg.APISecret) + "&" + url.QueryEscape(c.cfg.AccessSecret)

	// 生成签名
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	oauthParams["oauth_signature"] = signature

	// 构建 Authorization header
	var authPairs []string
	for k, v := range oauthParams {
		authPairs = append(authPairs, fmt.Sprintf(`%s="%s"`, url.QueryEscape(k), url.QueryEscape(v)))
	}
	sort.Strings(authPairs)
	authHeader := "OAuth " + strings.Join(authPairs, ", ")

	req.Header.Set("Authorization", authHeader)
}

func (c *client) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	// 添加更详细的错误提示
	hint := ""
	switch resp.StatusCode {
	case 401:
		hint = " (认证失败: 请检查 API Key/Secret 和 Access Token/Secret 是否正确)"
	case 403:
		hint = " (权限不足: 请检查 Twitter App 权限设置，确保已开启 Read/Write 权限，并重新生成 Access Token)"
	case 429:
		hint = " (请求过于频繁: 已触发速率限制，请稍后重试)"
	}

	return fmt.Errorf("twitter API error: status=%d%s, body=%s", resp.StatusCode, hint, string(body))
}

func (c *client) signRequest(req *http.Request, method, endpoint string, params map[string]string) {
	oauthParams := map[string]string{
		"oauth_consumer_key":     c.cfg.APIKey,
		"oauth_nonce":            uuid.New().String(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_token":            c.cfg.AccessToken,
		"oauth_version":          "1.0",
	}

	// Merge with additional params
	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range params {
		allParams[k] = v
	}

	// Parse URL params
	parsedURL, _ := url.Parse(endpoint)
	for k, v := range parsedURL.Query() {
		if len(v) > 0 {
			allParams[k] = v[0]
		}
	}

	// Sort keys
	var keys []string
	for k := range allParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build parameter string
	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, url.QueryEscape(k)+"="+url.QueryEscape(allParams[k]))
	}
	paramString := strings.Join(paramPairs, "&")

	// Build base URL (without query params)
	baseEndpoint := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path

	// Build signature base string
	signatureBase := method + "&" + url.QueryEscape(baseEndpoint) + "&" + url.QueryEscape(paramString)

	// Build signing key
	signingKey := url.QueryEscape(c.cfg.APISecret) + "&" + url.QueryEscape(c.cfg.AccessSecret)

	// Generate signature
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	oauthParams["oauth_signature"] = signature

	// Build Authorization header
	var authPairs []string
	for k, v := range oauthParams {
		authPairs = append(authPairs, fmt.Sprintf(`%s="%s"`, url.QueryEscape(k), url.QueryEscape(v)))
	}
	sort.Strings(authPairs)
	authHeader := "OAuth " + strings.Join(authPairs, ", ")

	req.Header.Set("Authorization", authHeader)
}
