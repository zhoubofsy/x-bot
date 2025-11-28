package dto

// WorkflowParams 工作流执行参数
type WorkflowParams struct {
	TweetCount int  `json:"tweet_count" form:"tweet_count"` // 每个用户获取的推文数量
	DryRun     bool `json:"dry_run" form:"dry_run"`         // 是否仅模拟执行
}

// WorkflowResult 工作流执行结果
type WorkflowResult struct {
	TotalUsers        int      `json:"total_users"`
	TotalTweets       int      `json:"total_tweets"`
	HackathonTweets   int      `json:"hackathon_tweets"`
	SuccessfulReplies int      `json:"successful_replies"`
	FailedReplies     int      `json:"failed_replies"`
	SkippedTweets     int      `json:"skipped_tweets"`
	Errors            []string `json:"errors,omitempty"`
}

// ProcessResult 单条推文处理结果
type ProcessResult struct {
	TweetID     string
	IsHackathon bool
	Success     bool
	Skipped     bool
	Error       error
}

// SyncFollowingResult 同步关注用户结果
type SyncFollowingResult struct {
	TotalCount   int      `json:"total_count"`
	NewCount     int      `json:"new_count"`
	UpdatedCount int      `json:"updated_count"`
	Errors       []string `json:"errors,omitempty"`
}

