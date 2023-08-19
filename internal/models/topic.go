package models

import (
	"gorm.io/gorm"
)

type TopicBlock struct {
	Hash      string   `json:"hash"`
	Title     string   `json:"title"`
	Creator   string   `json:"creator"`
	CID       string   `json:"cid"`
	Category  uint     `json:"category"`
	Tags      []uint   `json:"tags"`
	Images    []string `json:"images"`
	Upvotes   []string `json:"upvotes"`
	Downvotes []string `json:"downvotes"`
}

type Topic struct {
	gorm.Model
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
}
