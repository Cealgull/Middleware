package models

import (
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/utils"
	"gorm.io/gorm"
)

type Profile struct {
	ID             uint `gorm:"primaryKey"`
	Signature      string
	Credibility    uint
	Balance        int
	UserWallet     string
	User           *User    `gorm:"references:Wallet"`
	RolesAssigned  []*Role  `gorm:"foreignKey:OwnerID"`
	BadgesReceived []*Badge `gorm:"foreignKey:OwnerID"`
}

func (p *Profile) MarshalJSON() ([]byte, error) {

	type ProfileBadge struct {
		Name string `json:"name"`
		CID  string `json:"cid"`
	}

	return json.Marshal(&struct {
		Username  string `json:"username"`
		Wallet    string `json:"wallet"`
		Avatar    string `json:"avatar"`
		Signature string `json:"signature"`
		Muted     bool   `json:"muted"`
		Banned    bool   `json:"banned"`

		Balance     int  `json:"balance"`
		Credibility uint `json:"credibility"`

		ActiveRole     string          `json:"currentRole"`
		RolesAssigned  []string        `json:"rolesAssigned"`
		ActiveBadge    *ProfileBadge   `json:"currentBadge"`
		BadgesReceived []*ProfileBadge `json:"badgesReceived"`

		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}{

		Username:  p.User.Username,
		Wallet:    p.User.Wallet,
		Signature: p.Signature,
		Muted:     p.User.Muted,
		Banned:    p.User.Banned,

		Balance:     p.Balance,
		Credibility: p.Credibility,

		ActiveRole: func() string {
			if p.User.ActiveRole == nil {
				return ""
			} else {
				return p.User.ActiveRole.Name
			}
		}(),

		RolesAssigned: utils.Map(p.RolesAssigned, func(r *Role) string {
			return r.Name
		}),

		ActiveBadge: func() *ProfileBadge {
			if p.User.ActiveBadge == nil {
				return nil
			} else {
				return &ProfileBadge{
					p.User.ActiveBadge.Name,
					p.User.ActiveBadge.CID,
				}
			}
		}(),

		BadgesReceived: utils.Map(p.BadgesReceived, func(b *Badge) *ProfileBadge {
			return &ProfileBadge{
				b.Name,
				b.CID,
			}
		}),

		CreatedAt: p.User.CreatedAt,
		UpdatedAt: p.User.UpdatedAt,
	})

}

type User struct {
	gorm.Model

	Username string `gorm:"not null"`
	Wallet   string `gorm:"index:,unique,not null"`
	Avatar   string
	Muted    bool `gorm:"not null"`
	Banned   bool `gorm:"not null"`

	ActiveBadge *Badge `gorm:"foreignKey:OwnerID"`
	ActiveRole  *Role  `gorm:"foreignKey:OwnerID"`
}

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Username string `json:"username"`
		Wallet   string `json:"wallet"`
		Avatar   string `json:"avatar"`
		Muted    bool   `json:"muted"`
		Banned   bool   `json:"banned"`
		Badge    string `json:"badge"`
		Role     string `json:"role"`
	}{
		Username: u.Username,
		Wallet:   u.Wallet,
		Avatar:   u.Avatar,
		Muted:    u.Muted,
		Banned:   u.Banned,
		Badge:    u.ActiveBadge.CID,
		Role:     u.ActiveRole.Name,
	})
}

type ProfileBlock struct {
	Username  string `json:"username"`
	Wallet    string `json:"wallet"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	Muted     bool   `json:"muted"`
	Banned    bool   `json:"banned"`

	Balance     int  `json:"balance"`
	Credibility uint `json:"credibility"`

	ActiveRole     uint   `json:"activeRole"`
	RolesAssigned  []uint `json:"rolesAssigned"`
	ActiveBadge    uint   `json:"activeBadge"`
	BadgesReceived []uint `json:"badgesReceived"`
}

type Role struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	OwnerID     uint   `gorm:"index"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Badge struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	OwnerID     uint   `gorm:"index"`
	CID         string `json:"cid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
