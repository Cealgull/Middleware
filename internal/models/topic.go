package models

import (
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/utils"
	"gorm.io/gorm"
)

type TopicBlock struct {
	Hash     string   `json:"hash"`
	Title    string   `json:"title"`
	Creator  string   `json:"creator"`
	CID      string   `json:"cid"`
	Category string   `json:"category"`
	Tags     []uint   `json:"tags"`
	Images   []string `json:"images"`
}

type Topic struct {
	gorm.Model
	Hash          string `gorm:"uniqueIndex"`
	Title         string `gorm:"not null"`
	CreatorWallet string
	Creator       *User       `gorm:"references:Wallet"`
	Content       string      `gorm:"not null"`
	Category      string      `gorm:"not null"`
	Tags          []*TopicTag `gorm:"foreignkey:ID"`
	Assets        []*Asset    `gorm:"foreignkey:CID"`
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
	ID            uint   `gorm:"primaryKey" json:"-"`
	Name          string `gorm:"not null"   json:"name"`
	CreatorWallet string `gorm:"not null"`
	Creator       *User  `gorm:"references:Wallet"`
	Description   string `json:"description"`
}
