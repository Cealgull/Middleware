package models

type Community struct {
	ID         uint   `gorm:"primary_key"`
	Identifier string `gorm:"not null;unique"`
	Name       string `gorm:"not null;unique"`
  Description string `gorm:"not null"`
}
