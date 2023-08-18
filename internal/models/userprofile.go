package models

import (
	"encoding/json"
	"time"

	"github.com/Cealgull/Middleware/internal/utils"
	"gorm.io/gorm"
)

type Profile struct {
	ID                     uint `gorm:"primaryKey"`
	Signature              string
	Credibility            uint
	Balance                int
	UserWallet             string
	User                   *User            `gorm:"references:Wallet"`
	RoleRelationsAssigned  []*RoleRelation  `gorm:"polymorphic:Owner"`
	BadgeRelationsReceived []*BadgeRelation `gorm:"polymorphic:Owner"`
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
		Privilege   uint `json:"privilege"`

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
		Privilege: func() uint {
			if p.User.ActiveRoleRelation == nil {
				return 0
			} else {
				return p.User.ActiveRoleRelation.Role.Privilege
			}
		}(),

		ActiveRole: func() string {
			if p.User.ActiveRoleRelation == nil {
				return ""
			} else {
				return p.User.ActiveRoleRelation.Role.Name
			}
		}(),

		RolesAssigned: utils.Map(p.RoleRelationsAssigned, func(r *RoleRelation) string {
			return r.Role.Name
		}),

		ActiveBadge: func() *ProfileBadge {
			if p.User.ActiveBadgeRelation == nil {
				return nil
			} else {
				return &ProfileBadge{
					p.User.ActiveBadgeRelation.Badge.Name,
					p.User.ActiveBadgeRelation.Badge.CID,
				}
			}
		}(),

		BadgesReceived: utils.Map(p.BadgeRelationsReceived, func(b *BadgeRelation) *ProfileBadge {
			return &ProfileBadge{
				b.Badge.Name,
				b.Badge.CID,
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

	ActiveBadgeRelation *BadgeRelation `gorm:"polymorphic:Owner"`
	ActiveRoleRelation  *RoleRelation  `gorm:"polymorphic:Owner"`
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
		Badge: func(relation *BadgeRelation) string {
			if relation == nil {
				return "null"
			} else {
				return relation.Badge.CID
			}
		}(u.ActiveBadgeRelation),
		Role: func(relation *RoleRelation) string {
			if relation == nil {
				return "null"
			} else {
				return relation.Role.Name
			}
		}(u.ActiveRoleRelation),
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

type RoleRelation struct {
	ID        uint   `gorm:"primaryKey"`
	OwnerID   uint   `gorm:"not null"`
	OwnerType string `gorm:"not null"`
	RoleID    uint
	Role      *Role
}

type BadgeRelation struct {
	ID        uint   `gorm:"primaryKey"`
	OwnerID   uint   `gorm:"not null"`
	OwnerType string `gorm:"not null"`
	BadgeID   uint
	Badge     *Badge
}

type Role struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Privilege   uint   `json:"priledge"`
}

type Badge struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	CID         string `json:"cid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
