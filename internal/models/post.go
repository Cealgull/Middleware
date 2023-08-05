package models

import (
	"encoding/json"
	"time"
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
	ID       uint `gorm:"primaryKey"`
	Hash     string
	Creator  *User
	Content  string
	CreateAt time.Time
	UpdateAt time.Time
	ReplyTo  *Post
	BelongTo *Topic
	Assets   []*Asset
}

func (p *Post) MarshalJSON() ([]byte, error) {

	type DisplayReply struct {
		Creator  *User     `json:"creator"`
		Content  string    `json:"content"`
		UpdateAt time.Time `json:"updateAt"`
		Assets   []*Asset  `json:"assets,omitempty"`
	}

	return json.Marshal(&struct {
		ID       uint          `json:"id"`
		Hash     string        `json:"hash"`
		Creator  string        `json:"creator"`
		Content  string        `json:"content"`
		CreateAt time.Time     `json:"createAt"`
		UpdateAt time.Time     `json:"updateAt"`
		ReplyTo  *DisplayReply `json:"replyTo"`
		Assets   []*Asset      `json:"assets,omitempty"`
	}{
		ID:       p.ID,
		Hash:     p.Hash,
		Creator:  p.Creator.Username,
		Content:  p.Content,
		CreateAt: p.CreateAt,
		UpdateAt: p.UpdateAt,
		ReplyTo: &DisplayReply{
			Creator:  p.ReplyTo.Creator,
			Content:  p.Content,
			UpdateAt: p.ReplyTo.UpdateAt,
			Assets:   p.ReplyTo.Assets,
		},
		Assets: p.Assets,
	})
}
