package models

import (
	"encoding/json"
	"time"

)

type Asset struct {
  ID           uint   `gorm:"primaryKey"`
	CID           string `gorm:"index;not null" json:"cid"`
	CreatorWallet string
	Creator       *User  `gorm:"references:Wallet"`
	ContentType   string `gorm:"not null" json:"contentType"`
	OwnerID       uint
	OwnerType     string
  CreatedAt     time.Time      `gorm:"autoCreateTime"`
  UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
}

func (a *Asset) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Creator     string    `json:"creator"`
		ContentType string    `json:"contentType"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		CID         string    `json:"cid"`
	}{
		ContentType: a.ContentType,
		Creator:     a.CreatorWallet,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		CID:         a.CID,
	})
}
