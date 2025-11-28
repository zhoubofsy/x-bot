package entity

import "time"

type BotConfig struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	ConfigKey   string    `json:"config_key" gorm:"column:config_key;uniqueIndex;size:128;not null"`
	ConfigValue string    `json:"config_value" gorm:"column:config_value;type:text;not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (BotConfig) TableName() string {
	return "bot_configs"
}

