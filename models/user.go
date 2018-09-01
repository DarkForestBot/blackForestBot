package models

import (
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// User is used in database
type User struct {
	ID         int
	TgUserID   int64  `gorm:"unique;not null;size:50"`
	TgUserName string `gorm:"key;not null;size:255"`
	Name       string `gorm:"not null"`
	Language   string `gorm:"default:\"English\""`

	//User stats
	GamesJoined         int `gorm:"default:0;not null"`
	GamesJoinedAchive   int `gorm:"default:0;not null"`
	GamesWon            int `gorm:"default:0;not null"`
	GamesWonAchive      int `gorm:"default:0;not null"`
	ShootCount          int `gorm:"default:0;not null"`
	ShootAchive         int `gorm:"default:0;not null"`
	BetrayCount         int `gorm:"default:0;not null"`
	BetrayAchive        int `gorm:"default:0;not null"`
	KillCount           int `gorm:"default:0;not null"`
	KillAchive          int `gorm:"default:0;not null"`
	TrapCount           int `gorm:"defalut:0;not null"`
	TrapAchive          int `gorm:"default:0;not null"`
	UnionCount          int `gorm:"default:0;not null"`
	UnionAchive         int `gorm:"default:0;not null"`
	UnionSuccessCount   int `gorm:"default:0;not null"`
	UnionSuccessAchive  int `gorm:"default:0;not null"`
	BeUnionedCount      int `gorm:"default:0;not null"`
	BeUnionedAchive     int `gorm:"default:0;not null"`
	AchiveRewardedCount int `gorm:"default:0;not null"`

	//Wont record into database below
	QueryMsg        *tgApi.Message `gorm:"-"`
	TgGroupJoinGame *TgGroup       `gorm:"-"`
}

//Update is
func (u *User) Update() error {
	return database.DB.Save(u).Error
}

//Stats is
func (u *User) Stats(to *tgApi.Message) {
	u.QueryMsg = to
	UserStatsHint <- u
}

//GetUser is
func GetUser(tgID int64) (*User, error) {
	user := new(User)
	if err := database.DB.Where(User{TgUserID: tgID}).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

//GetUserByUserName is
func GetUserByUserName(name string) (*User, error) {
	user := new(User)
	if err := database.DB.Where(User{TgUserName: name}).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
