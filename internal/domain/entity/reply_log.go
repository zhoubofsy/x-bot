package entity

import "time"

type ReplyStatus string

const (
	ReplyStatusPending ReplyStatus = "pending"
	ReplyStatusSuccess ReplyStatus = "success"
	ReplyStatusFailed  ReplyStatus = "failed"
	ReplyStatusSkipped ReplyStatus = "skipped"
	ReplyStatusDryRun  ReplyStatus = "dry_run"
)

type ReplyLog struct {
	ID            int         `json:"id" gorm:"primaryKey"`
	TweetID       string      `json:"tweet_id" gorm:"column:tweet_id;uniqueIndex;size:64;not null"`
	TweetAuthorID string      `json:"tweet_author_id" gorm:"column:tweet_author_id;size:64;not null"`
	TweetContent  string      `json:"tweet_content" gorm:"type:text"`
	ReplyTweetID  string      `json:"reply_tweet_id" gorm:"column:reply_tweet_id;size:64"`
	AdCopyID      *int        `json:"ad_copy_id" gorm:"column:ad_copy_id"`
	AdCopy        *AdCopy     `json:"ad_copy,omitempty" gorm:"foreignKey:AdCopyID"`
	Status        ReplyStatus `json:"status" gorm:"size:32;default:pending;index"`
	ErrorMessage  string      `json:"error_message" gorm:"type:text"`
	LLMResponse   string      `json:"llm_response" gorm:"column:llm_response;type:text"`
	IsHackathon   bool        `json:"is_hackathon" gorm:"column:is_hackathon"`
	CreatedAt     time.Time   `json:"created_at" gorm:"index"`
}

func (ReplyLog) TableName() string {
	return "reply_logs"
}

