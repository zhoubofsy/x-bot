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
	baseURL  = "https://api.twitter.com/2"
	oauthURL = "https://api.twitter.com"
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

	c.signRequest(req, "POST", endpoint, nil)

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

func (c *client) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("twitter API error: status=%d, body=%s", resp.StatusCode, string(body))
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
