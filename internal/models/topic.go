package models

import (
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/utils"
	"gorm.io/gorm"
)

type DeleteBlock struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

type TopicBlock struct {
	Hash     string   `json:"hash"`
	Title    string   `json:"title"`
	Creator  string   `json:"creator"`
	CID      string   `json:"cid"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Images   []string `json:"images"`

	Deleted bool `json:"deleted"`

	Upvotes   []string            `json:"upvotes"`
	Downvotes []string            `json:"downvotes"`
	Emojis    map[string][]string `json:"emojis"`
}

type Topic struct {
	ID               uint   `gorm:"primaryKey"`
	Hash             string `gorm:"uniqueIndex;not null"`
	Title            string `gorm:"not null"`
	CreatorWallet    string `gorm:"index;not null"`
	Creator          *User  `gorm:"references:Wallet"`
	Content          string `gorm:"not null"`
	CategoryAssigned *CategoryRelation
	TagsAssigned     []*TagRelation `gorm:"polymorphic:Owner"`
	Upvotes          []*Upvote      `gorm:"polymorphic:Owner"`
	Downvotes        []*Downvote    `gorm:"polymorphic:Owner"`
	Assets           []*Asset       `gorm:"polymorphic:Owner"`
	Closed           bool           `gorm:"not null"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (t *Topic) MarshalJSON() ([]byte, error) {

	type DisplayTag struct {
		Name string `json:"name"`
	}

	type DisplayCategory struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	type DisplayAsset struct {
		Creator     string    `json:"creator"`
		ContentType string    `json:"contentType"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		CID         string    `json:"cid"`
	}

	return json.Marshal(&struct {
		Hash             string           `json:"hash"`
		Title            string           `json:"title"`
		Creator          *User            `json:"creator"`
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
		Hash:    t.Hash,
		Title:   t.Title,
		Creator: t.Creator,
		Content: t.Content,
		CategoryAssigned: func() *DisplayCategory {
			if t.CategoryAssigned != nil {
				return &DisplayCategory{Name: t.CategoryAssigned.CategoryName,
					Color: t.CategoryAssigned.Category.Color}
			}
			return nil
		}(),
		TagsAssigned: utils.Map(t.TagsAssigned, func(t *TagRelation) *DisplayTag {
			return &DisplayTag{Name: t.TagName}
		}),
		Upvotes:   utils.Map(t.Upvotes, func(u *Upvote) string { return u.CreatorWallet }),
		Downvotes: utils.Map(t.Downvotes, func(d *Downvote) string { return d.CreatorWallet }),
		Assets:    t.Assets,
		Closed:    t.Closed,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	})
}
