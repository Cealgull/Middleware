package models

import (
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/utils"
	"gorm.io/gorm"
)

type TopicBlock struct {
	Hash      string   `json:"hash"`
	Title     string   `json:"title"`
	Creator   string   `json:"creator"`
	CID       string   `json:"cid"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	Images    []string `json:"images"`
	Upvotes   []string `json:"upvotes"`
	Downvotes []string `json:"downvotes"`
}

type Topic struct {
	ID               uint   `gorm:"primaryKey"`
	Hash             string `gorm:"not null"`
	Title            string `gorm:"not null"`
	CreatorWallet    string `gorm:"not null"`
	Creator          *User  `gorm:"references:Wallet"`
	Content          string `gorm:"not null"`
	CategoryAssigned *CategoryRelation
	TagsAssigned     []*TagRelation `gorm:"polymorphic:Owner"`
	Upvotes          []*Upvote      `gorm:"polymorphic:Owner"`
	Downvote         []*Downvote    `gorm:"polymorphic:Owner"`
	Assets           []*Asset       `gorm:"polymorphic:Owner"`
	Closed           bool           `gorm:"not null"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (t *Topic) MarshalJSON() ([]byte, error) {

	type DisplayTag struct {
		Name string `json:"title"`
    Color string `json:"color"`
	}

	type DisplayCategory struct {
		Name  string `json:"title"`
		Color uint   `json:"color"`
	}

	type DisplayAsset struct {
		Creator     string    `json:"creator"`
		ContentType string    `json:"contentType"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		CID         string    `json:"cid"`
	}

	return json.Marshal(&struct {
		ID               uint             `json:"id"`
		Hash             string           `json:"hash"`
		Title            string           `json:"title"`
		Creator          *User            `json:"creator"`
		Avatar           string           `json:"avatar"`
		Content          string           `json:"content"`
		CategoryAssigned *DisplayCategory `json:"categoryAssigned"`
		TagsAssigned     []*DisplayTag    `json:"tagsAssigned"`
		Upvotes          []string         `json:"upvotes"`
		Downvotes        []string         `json:"downvotes"`
		Assets           []*Asset         `json:"assets"`
		Closed           bool             `json:"closed"`
		CreatedAt        time.Time        `json:"createdAt"`
		UpdatedAt        time.Time        `json:"updatedAt"`
	}{
		ID:               t.ID,
		Hash:             t.Hash,
		Title:            t.Title,
		Creator:          t.Creator,
		Content:          t.Content,
		CategoryAssigned: &DisplayCategory{Name: t.CategoryAssigned.Category.Name},
		TagsAssigned: utils.Map(t.TagsAssigned, func(t *TagRelation) *DisplayTag {
        return &DisplayTag{ Name: t.Tag.Name}
		}),
		Upvotes:   utils.Map(t.Upvotes, func(u *Upvote) string { return u.Creator.Wallet }),
		Downvotes: utils.Map(t.Downvote, func(d *Downvote) string { return d.Creator.Wallet }),
		Assets:    t.Assets,
		Closed:    t.Closed,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	})
}
