package models

//GameMode is
type GameMode int

const (
	Normal GameMode = 0
	Expert GameMode = 1
)

// TgGroup is used in database
type TgGroup struct {
	ID        int
	TgGroupID int64  `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	AdminID   int
	Admin     User
	Lang      string   `gorm:"default:\"English\""`
	Mode      GameMode `gorm:"default:0"`
	Active    bool     `gorm:"default:1"`
}
