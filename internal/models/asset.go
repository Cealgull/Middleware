package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Asset struct {
	gorm.Model
	CreatorWallet string
	Creator       *User  `gorm:"references:Wallet"`
	ContentType   string `gorm:"not null" json:"contentType"`
	CID           string `gorm:"uniqueIndex,not null" json:"cid"`
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
		Creator:     a.Creator.Username,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		CID:         a.CID,
	})
}
