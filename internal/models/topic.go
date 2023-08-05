package models

import (
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/utils"
	"gorm.io/gorm"
)

type TopicBlock struct {
	Hash      string    `json:"hash"`
	Title     string    `json:"title"`
	Creator   string    `json:"creator"`
	CID       string    `json:"cid"`
	CreateAt  time.Time `json:"createAt"`
	UpdateAt  time.Time `json:"updateAt"`
	DeletedAt time.Time `json:"deletedAt"`
	Category  string    `json:"category"`
	Tags      []uint    `json:"tags"`
	Images    []string  `json:"images"`
}

type Topic struct {
	gorm.Model
	Hash     string `gorm:"index"`
	Title    string
	Wallet   string `gorm:"index" json:"-"`
	Creator  *User  `gorm:"foreignKey:Wallet,references:Wallet"`
	Content  string
	Category string
	Tags     []*TopicTag
	Assets   []*Asset
}

func (t *Topic) MarshalJSON() ([]byte, error) {

	type DisplayTag struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	return json.Marshal(&struct {
		ID       uint          `json:"id"`
		Hash     string        `json:"hash"`
		Title    string        `json:"title"`
		Creator  string        `json:"creator"`
		Content  string        `json:"content"`
		CreateAt time.Time     `json:"createAt"`
		UpdateAt time.Time     `json:"updateAt"`
		Category string        `json:"category"`
		Tags     []*DisplayTag `json:"tags"`
		Assets   []*Asset      `json:"assets"`
	}{
		ID:       t.ID,
		Hash:     t.Hash,
		Title:    t.Title,
		Creator:  t.Creator.Username,
		Content:  t.Content,
		CreateAt: t.CreatedAt,
		UpdateAt: t.UpdatedAt,
		Category: t.Category,
		Tags:     utils.Map(t.Tags, func(t *TopicTag) *DisplayTag { return &DisplayTag{ID: t.ID, Name: t.Name} }),
		Assets:   t.Assets,
	})
}

type TopicTag struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	Name        string `gorm:"index" json:"name"`
	Creator     *User  `gorm:"foreignKey:ID"`
	Description string `json:"description"`
}
