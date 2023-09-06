package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID            uint   `gorm:"primaryKey" json:"-"`
	Name          string `gorm:"uniqueIndex;not null"   json:"name"`
	CreatorWallet string `gorm:"not null"`
	Creator       *User  `gorm:"foreignKey:CreatorWallet;references:Wallet"`
	Description   string `json:"description"`
}

type TagBlock struct {
	Name          string `json:"name"`
	CreatorWallet string `json:"creatorWallet"`
	Description   string `json:"description"`
}

type TagRelation struct {
	ID        uint   `gorm:"primaryKey"`
	OwnerID   uint   `gorm:"not null"`
	OwnerType string `gorm:"not null"`
	TagName   string `gorm:"not null"`
	Tag       *Tag   `gorm:"foreignKey:TagName;references:Name"`
}

type Category struct {
	ID                uint   `gorm:"primaryKey"`
	CategoryGroupName string `gorm:"not null"`
	Color             uint   `gorm:"not null"`
	Name              string `gorm:"uniqueIndex;not null"`
}

type CategoryBlock struct {
	CategoryGroupName string `json:"categoryGroupName"`
	Color             uint   `json:"color"`
	Name              string `json:"name"`
}

type CategoryRelation struct {
	ID           uint           `gorm:"primaryKey"`
	TopicID      uint           `gorm:"not null"`
	CategoryName string         `gorm:"not null"`
	Category     *Category      `gorm:"foreignKey:CategoryName;references:Name"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type CategoryGroup struct {
	ID         uint           `gorm:"primaryKey"`
	Name       string         `gorm:"uniqueIndex;not null"`
	Color      uint           `gorm:"not null"`
	Categories []*Category    `gorm:"foreignKey:CategoryGroupName;references:Name"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type CategoryGroupBlock struct {
	Name       string   `json:"name"`
	Color      uint     `json:"color"`
	Categories []string `json:"categories"`
}

type Upvote struct {
	ID            uint           `gorm:"primaryKey"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatorWallet string         `gorm:"index;not null"`
	Creator       *User          `gorm:"foreignKey:CreatorWallet;references:Wallet"`
	OwnerID       uint
	OwnerType     string
}

type UpvoteBlock struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

type Downvote struct {
	ID            uint           `gorm:"primaryKey"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatorWallet string         `gorm:"index:idx_wallet;not null"`
	Creator       *User          `gorm:"foreignKey:CreatorWallet;references:Wallet"`
	OwnerID       uint
	OwnerType     string
}

type DownvoteBlock struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
}

type EmojiRelation struct {
	ID        uint           `gorm:"primaryKey"`
	EmojiCode string         `gorm:"index:,unique;not null"`
	Emoji     *Emoji         `gorm:"foreignKey:EmojiCode;references:Code"`
	OwnerID   uint           `gorm:"not null"`
	OwnerType string         `gorm:"not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type EmojiBlock struct {
	Hash    string `json:"hash"`
	Creator string `json:"creator"`
	Code    string `json:"code"`
}

type Emoji struct {
	ID   uint   `gorm:"primaryKey"`
	Code string `gorm:"not null"`
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
