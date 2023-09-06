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
	UserWallet             *string          `gorm:"index:,unique,sort:desc,not null"`
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

		ActiveRole     string          `json:"activeRole"`
		RolesAssigned  []string        `json:"rolesAssigned"`
		ActiveBadge    *ProfileBadge   `json:"activeBadge"`
		BadgesReceived []*ProfileBadge `json:"badgesReceived"`

		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}{

		Username:  p.User.Username,
		Avatar:    p.User.Avatar,
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
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"not null"`
	Wallet    string `gorm:"index:,unique,sort:desc,not null"`
	Avatar    string
	Muted     bool           `gorm:"not null"`
	Banned    bool           `gorm:"not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ActiveBadgeRelation *BadgeRelation `gorm:"polymorphic:Owner"`
	ActiveRoleRelation  *RoleRelation  `gorm:"polymorphic:Owner"`
}

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Username string `json:"username"`
		Wallet   string `json:"wallet"`
		Avatar   string `json:"avatar"`
		Badge    string `json:"badge"`
		Role     string `json:"role"`
	}{
		Username: u.Username,
		Wallet:   u.Wallet,
		Avatar:   u.Avatar,
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

	ActiveRole     string   `json:"activeRole"`
	RolesAssigned  []string `json:"rolesAssigned"`
	ActiveBadge    string   `json:"activeBadge"`
	BadgesReceived []string `json:"badgesReceived"`
}

type RoleRelation struct {
	ID        uint   `gorm:"primaryKey"`
	OwnerID   uint   `gorm:"not null"`
	OwnerType string `gorm:"not null"`
	RoleName  string
	Role      *Role `gorm:"foreignKey:RoleName;references:Name"`
}

type BadgeRelation struct {
	ID        uint   `gorm:"primaryKey"`
	OwnerID   uint   `gorm:"not null"`
	OwnerType string `gorm:"not null"`
	BadgeName string
	Badge     *Badge `gorm:"foreignKey:BadgeName;references:Name"`
}

type Role struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	Name        string `gorm:"uniqueIndex" json:"name"`
	Description string `json:"description"`
	Privilege   uint   `json:"priledge"`
}

type Badge struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	CID         string `json:"cid"`
	Name        string `gorm:"uniqueIndex" json:"name"`
	Description string `json:"description"`
}
