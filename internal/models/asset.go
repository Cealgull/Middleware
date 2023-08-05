package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Asset struct {
	gorm.Model
	Creator     User   `gorm:"foreignKey:ID"`
	ContentType string `json:"contentType"`
	CID         string `json:"cid"`
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
