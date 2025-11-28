package entity

import "time"

type AdCopy struct {
	ID         int        `json:"id" gorm:"primaryKey"`
	Name       string     `json:"name" gorm:"size:128;not null"`
	Content    string     `json:"content" gorm:"type:text;not null"`
	Category   string     `json:"category" gorm:"size:64;default:hackathon;index"`
	Priority   int        `json:"priority" gorm:"default:0"`
	IsActive   bool       `json:"is_active" gorm:"default:true;index"`
	UseCount   int        `json:"use_count" gorm:"default:0"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (AdCopy) TableName() string {
	return "ad_copies"
}

type CreateAdCopyInput struct {
	Name     string `json:"name" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Category string `json:"category"`
	Priority int    `json:"priority"`
}

type UpdateAdCopyInput struct {
	Name     *string `json:"name"`
	Content  *string `json:"content"`
	Category *string `json:"category"`
	Priority *int    `json:"priority"`
	IsActive *bool   `json:"is_active"`
}

