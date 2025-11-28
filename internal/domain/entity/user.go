package entity

import "time"

type FollowedUser struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	TwitterUserID string    `json:"twitter_user_id" gorm:"column:twitter_user_id;uniqueIndex;size:64"`
	Username      string    `json:"username" gorm:"size:128"`
	DisplayName   string    `json:"display_name" gorm:"size:256"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (FollowedUser) TableName() string {
	return "followed_users"
}

