package twitter

import "time"

// TwitterUser Twitter用户信息
type TwitterUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Tweet 推文信息
type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
}

// APIResponse Twitter API 响应
type APIResponse struct {
	Data     interface{} `json:"data"`
	Includes *Includes   `json:"includes,omitempty"`
	Meta     *Meta       `json:"meta,omitempty"`
	Errors   []APIError  `json:"errors,omitempty"`
}

type Includes struct {
	Users []TwitterUser `json:"users,omitempty"`
}

type Meta struct {
	ResultCount   int    `json:"result_count"`
	NextToken     string `json:"next_token,omitempty"`
	PreviousToken string `json:"previous_token,omitempty"`
}

type APIError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Type   string `json:"type"`
}

// CreateTweetRequest 创建推文请求
type CreateTweetRequest struct {
	Text  string       `json:"text"`
	Reply *ReplyConfig `json:"reply,omitempty"`
}

type ReplyConfig struct {
	InReplyToTweetID string `json:"in_reply_to_tweet_id"`
}

// CreateTweetResponse 创建推文响应
type CreateTweetResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

// FollowingResponse 关注列表响应
type FollowingResponse struct {
	Data []TwitterUser `json:"data"`
	Meta *Meta         `json:"meta,omitempty"`
}

// TweetsResponse 推文列表响应
type TweetsResponse struct {
	Data []Tweet `json:"data"`
	Meta *Meta   `json:"meta,omitempty"`
}

