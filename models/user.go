package models

import (
	"git.wetofu.top/tonychee7000/blackForestBot/database"
)

// User is used in database
type User struct {
	ID                  int
	TgUserID            int64  `gorm:"unique;not null;size:50"`
	TgUserName          string `gorm:"not null;size:255"`
	Name                string `gorm:"not null"`
	GamesJoined         int    `gorm:"default:0;not null"`
	GamesWon            int    `gorm:"default:0;not null"`
	Language            string `gorm:"default:\"English\""`
	ShootCount          int    `gorm:"default:0;not null"`
	BetrayCount         int    `gorm:"default:0;not null"`
	KillCount           int    `gorm:"default:0;not null"`
	TrapCount           int    `gorm:"defalut:0;not null"`
	UnionCount          int    `gorm:"default:0;not null"`
	UnionSuccessCount   int    `gorm:"default:0;not null"`
	BeUnionedCount      int    `gorm:"default:0;not null"`
	AchiveRewardedCount int    `gorm:"default:0;not null"`
}

//Update is
func (u *User) Update() error {
	return database.DB.Save(u).Error
}

//GetUser is
func GetUser(tgID int64) (*User, error) {
	user := new(User)
	if err := database.DB.Where(User{TgUserID: tgID}).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
