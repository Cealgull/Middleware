package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	Name        string `gorm:"not null"   json:"name"`
	CreatorID   uint   `gorm:"not null"`
	Creator     *User
	Description string `json:"description"`
}

type TagBlock struct {
	Name        string `json:"name"`
	CreatorID   uint   `json:"creatorID"`
	Description string `json:"description"`
}

type TagRelation struct {
	ID        uint   `gorm:"primaryKey"`
	OwnerID   uint   `gorm:"not null"`
	OwnerType string `gorm:"not null"`
	TagID     uint   `gorm:"not null"`
	Tag       *Tag
}

type Category struct {
	ID              uint   `gorm:"primaryKey"`
	CategoryGroupID uint   `gorm:"not null"`
	Color           uint   `gorm:"not null"`
	Name            string `gorm:"not null"`
}

type CategoryBlock struct {
	CategoryGroupID uint   `json:"categoryGroupID"`
	Color           uint   `json:"color"`
	Name            string `json:"name"`
}

type CategoryRelation struct {
	ID         uint `gorm:"primaryKey"`
	TopicID    uint `gorm:"not null"`
	CategoryID uint `gorm:"not null"`
	Category   *Category
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type CategoryGroup struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"not null"`
	Color      uint   `gorm:"not null"`
	Categories []*Category
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type CategoryGroupBlock struct {
	Name       string    `json:"name"`
	Color      uint      `json:"color"`
	Categories []uint    `json:"categories"`
	CreateAt   time.Time `json:"createAt"`
}

type Upvote struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatorID uint           `gorm:"not null"`
	Creator   *User
	OwnerID   uint
	OwnerType string
}

type Downvote struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatorID uint           `gorm:"not null"`
	Creator   *User
	OwnerID   uint
	OwnerType string
}

type EmojiRelation struct {
	ID        uint `gorm:"primaryKey"`
	EmojiID   uint `gorm:"not null"`
	Emoji     *Emoji
	OwnerID   uint           `gorm:"not null"`
	OwnerType string         `gorm:"not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Emoji struct {
	ID   uint `gorm:"primaryKey"`
	Code uint `gorm:"not null"`
}

func (u *Upvote) MarshalJSON() ([]byte, error) {

	return json.Marshal(&struct {
		Avatar   string `json:"avatar"`
		Username string `json:"username"`
	}{
		Avatar:   u.Creator.Avatar,
		Username: u.Creator.Username,
	})

}

func (d *Downvote) MarshalJSON() ([]byte, error) {

	return json.Marshal(&struct {
		Avatar   string `json:"avatar"`
		Username string `json:"username"`
	}{
		Avatar:   d.Creator.Avatar,
		Username: d.Creator.Username,
	})

}
