package entity

import "time"

// Tweet 推文实体（非数据库存储，用于业务传递）
type Tweet struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

