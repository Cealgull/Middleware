package models

import (
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/utils"
	"gorm.io/gorm"
)

type PostRequest struct {
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	ReplyTo  string   `json:"replyTo"`
	BelongTo string   `json:"belongTo"`
}

type PostBlock struct {
	Hash     string    `json:"hash"`
	Creator  string    `json:"creator"`
	CID      string    `json:"cid"`
	CreateAt time.Time `json:"createAt"`
	UpdateAt time.Time `json:"updateAt"`
	ReplyTO  string    `json:"replyTo"`
	BelongTo string    `json:"belongTo"`
	Assets   []*Asset  `json:"assets,omitempty"`
}

type Post struct {
	ID            uint   `gorm:"primaryKey"`
	Hash          string `gorm:"uniqueIndex"`
	CreatorWallet string
	Creator       *User     `gorm:"references:Wallet"`
	Content       string    `gorm:"not null"`
	CreateAt      time.Time `gorm:"not null"`
	UpdateAt      time.Time `gorm:"not null"`
	DeletedAt     gorm.DeletedAt
	ReplyTo       *Post `gorm:"foreignKey:ID"`
	BelongToID    uint  `gorm:"not null"`
	BelongTo      *Topic

	Upvotes   []*Upvote   `gorm:"polymorphic:Owner"`
	Downvotes []*Downvote `gorm:"polymorphic:Owner"`
	Closed    bool        `gorm:"not null"`
	Assets    []*Asset    `gorm:"polymorphic:Owner"`
}

func (p *Post) MarshalJSON() ([]byte, error) {

	type DisplayReply struct {
		Creator  *User     `json:"creator"`
    Hash     string    `json:"hash"`
		Content  string    `json:"content"`
		UpdateAt time.Time `json:"updateAt"`
		Assets   []*Asset  `json:"assets,omitempty"`
	}

	return json.Marshal(&struct {
		ID        uint          `json:"id"`
		Hash      string        `json:"hash"`
		Creator   *User         `json:"creator"`
		Content   string        `json:"content"`
		CreateAt  time.Time     `json:"createAt"`
		UpdateAt  time.Time     `json:"updateAt"`
		ReplyTo   *DisplayReply `json:"replyTo"`
		Assets    []*Asset      `json:"assets,omitempty"`
		Upvotes   []string      `json:"upvotes"`
		Downvotes []string      `json:"downvotes"`
		BelongTo  string        `json:"belongTo"`
	}{
		ID:       p.ID,
		Hash:     p.Hash,
		Creator:  p.Creator,
		Content:  p.Content,
		CreateAt: p.CreateAt,
		UpdateAt: p.UpdateAt,
		ReplyTo: &DisplayReply{
      Hash: p.ReplyTo.Hash,
			Creator:  p.ReplyTo.Creator,
			Content:  p.Content,
			UpdateAt: p.ReplyTo.UpdateAt,
			Assets:   p.ReplyTo.Assets,
		},
		BelongTo: func() string {
			if p.BelongTo != nil {
				return p.BelongTo.Title
			}
			return ""

		}(),
		Assets: p.Assets,
		Upvotes: utils.Map(p.Upvotes, func(upvote *Upvote) string {
			return upvote.Creator.Wallet
		}),
		Downvotes: utils.Map(p.Downvotes, func(downvote *Downvote) string {
			return downvote.Creator.Wallet
		}),
	})
}
