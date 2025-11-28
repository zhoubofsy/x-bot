package entity

import "time"

type FollowedUser struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	TwitterUserID string    `json:"twitter_user_id" gorm:"column:twitter_user_id;type:varchar(64);unique"`
	Username      string    `json:"username" gorm:"type:varchar(128)"`
	DisplayName   string    `json:"display_name" gorm:"type:varchar(256)"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (FollowedUser) TableName() string {
	return "followed_users"
}
