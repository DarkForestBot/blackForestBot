package models

import "git.wetofu.top/tonychee7000/blackForestBot/database"

//GameMode is
type GameMode int

//List of gamemode
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

//Update is
func (t *TgGroup) Update() error {
	return database.DB.Save(t).Error
}

//GetTgGroup is
func GetTgGroup(ID int64) (*TgGroup, error) {
	group := new(TgGroup)
	if err := database.DB.Where(TgGroup{TgGroupID: ID}).First(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}
