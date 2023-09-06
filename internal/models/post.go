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
	Hash     string   `json:"hash"`
	Creator  string   `json:"creator"`
	CID      string   `json:"cid"`
	ReplyTo  string   `json:"replyTo"`
	BelongTo string   `json:"belongTo"`
	Assets   []string `json:"assets,omitempty"`
}

type Post struct {
	ID            uint      `gorm:"primaryKey"`
	Hash          string    `gorm:"uniqueIndex"`
	CreatorWallet string    `gorm:"index;not null"`
	Creator       *User     `gorm:"references:Wallet"`
	Content       string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime;not null"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;not null"`
	DeletedAt     gorm.DeletedAt
	ReplyToID     *uint
	ReplyTo       *Post
	BelongToHash  string `gorm:"index;not null"`
	BelongTo      *Topic `gorm:"references:Hash;foreignKey:BelongToHash"`

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
    CreateAt time.Time `json:"createAt"`
		UpdateAt time.Time `json:"updateAt"`
		Assets   []*Asset  `json:"assets,omitempty"`
	}

	return json.Marshal(&struct {
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
		Hash:     p.Hash,
		Creator:  p.Creator,
		Content:  p.Content,
		CreateAt: p.CreatedAt,
		UpdateAt: p.UpdatedAt,
		ReplyTo: func() *DisplayReply {
			if p.ReplyTo != nil {
				return &DisplayReply{
					Hash:     p.ReplyTo.Hash,
					Creator:  p.ReplyTo.Creator,
					Content:  p.Content,
          CreateAt: p.ReplyTo.CreatedAt,
					UpdateAt: p.ReplyTo.UpdatedAt,
					Assets:   p.ReplyTo.Assets,
				}
			}
			return nil
		}(),
		BelongTo: func() string {
			if p.BelongTo != nil {
				return p.BelongTo.Title
			}
			return ""

		}(),
		Assets: p.Assets,
		Upvotes: utils.Map(p.Upvotes, func(upvote *Upvote) string {
			return upvote.CreatorWallet
		}),
		Downvotes: utils.Map(p.Downvotes, func(downvote *Downvote) string {
			return downvote.CreatorWallet
		}),
	})
}
