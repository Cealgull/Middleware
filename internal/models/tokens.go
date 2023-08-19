package models

type OwnedToken struct {
	ID           uint `gorm:"primaryKey"`
	OwnerID      uint
	Asset        *Asset         `gorm:"polymorphic:Owner;"`
	TagsAssigned []*TagRelation `gorm:"polymorphic:Owner;"`
	Volume       uint
}

type TradedToken struct {
	ID           uint `gorm:"primaryKey"`
	OwnerID      uint
	Owner        *User
	Asset        *Asset         `gorm:"polymorphic:Owner;"`
	TagsAssigned []*TagRelation `gorm:"polymorphic:Owner;"`
	Upvotes      []*Upvote      `gorm:"polymorphic:Owner;"`
	Downvotes    []*Downvote    `gorm:"polymorphic:Owner;"`
	Value        uint
	Volume       uint
}
