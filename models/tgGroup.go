package models

import (
	"fmt"

	"git.wetofu.top/tonychee7000/blackForestBot/database"
)

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
	Admin     *User
	Lang      string   `gorm:"default:\"English\""`
	Mode      GameMode `gorm:"default:0"`
	Active    bool     `gorm:"default:1"`
}

//Update is
func (t *TgGroup) Update() error {
	return database.DB.Set("gorm:save_associations", false).Save(t).Error
}

func (t *TgGroup) String() string {
	return fmt.Sprintf("TgGroup(TgGroupID=%d Name=`%s` Admin=%s Lang=`%s` Mode=%d Active=%v)",
		t.TgGroupID, t.Name, t.Admin, t.Lang, t.Mode, t.Active)
}

//GetTgGroup is
func GetTgGroup(ID int64) (*TgGroup, error) {
	group := new(TgGroup)
	user := new(User)
	if err := database.DB.Where(TgGroup{TgGroupID: ID}).First(group).Error; err != nil {
		return nil, err
	}
	if err := database.DB.Where(User{ID: group.AdminID}).First(user).Error; err != nil {
		return nil, err
	}
	group.Admin = user
	return group, nil
}
