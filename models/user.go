package models

import (
	"fmt"

	"git.wetofu.top/tonychee7000/blackForestBot/database"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const maxAchivementLevel = 5

//List of achivement code
const (
	AchivementGamesJoined  = iota // 0
	AchivementGamesWon            // 1
	AchivementShoot               // 2
	AchivementBetray              // 3
	AchivementKill                // 4
	AchivementGuessKill           // 5
	AchivementSniperKill          // 6
	AchivementKilledByTrap        // 7
	AchivementTrap                // 8
	AchivementUnion               // 9
	AchivementUnionSuccess        // 10
	AchivementBeUnioned           // 11
)

// User is used in database
type User struct {
	ID         int
	TgUserID   int64  `gorm:"unique;not null;size:50"`
	TgUserName string `gorm:"key;not null;size:255"`
	Name       string `gorm:"not null"`
	Language   string `gorm:"default:\"English\""`

	//User stats
	GamesJoined           int `gorm:"default:0;not null"`
	GamesJoinedAchive     int `gorm:"default:0;not null"`
	GamesWon              int `gorm:"default:0;not null"`
	GamesWonAchive        int `gorm:"default:0;not null"`
	ShootCount            int `gorm:"default:0;not null"`
	ShootAchive           int `gorm:"default:0;not null"`
	BetrayCount           int `gorm:"default:0;not null"`
	BetrayAchive          int `gorm:"default:0;not null"`
	KillCount             int `gorm:"default:0;not null"`
	KillAchive            int `gorm:"default:0;not null"`
	GuessKillCount        int `gorm:"default:0;not null"`
	GuessKillCountAchive  int `gorm:"default:0;not null"`
	SniperKillCount       int `gorm:"default:0;not null"`
	SniperKillCountAchive int `gorm:"default:0;not null"`
	KilledByTrapCount     int `gorm:"default:0;not null"`
	KilledByTrapAchive    int `gorm:"default:0;not null"`
	TrapCount             int `gorm:"defalut:0;not null"`
	TrapAchive            int `gorm:"default:0;not null"`
	UnionCount            int `gorm:"default:0;not null"`
	UnionAchive           int `gorm:"default:0;not null"`
	UnionSuccessCount     int `gorm:"default:0;not null"`
	UnionSuccessAchive    int `gorm:"default:0;not null"`
	BeUnionedCount        int `gorm:"default:0;not null"`
	BeUnionedAchive       int `gorm:"default:0;not null"`
	AchiveRewardedCount   int `gorm:"default:0;not null"`

	//Wont record into database below
	QueryMsg        *tgApi.Message `gorm:"-"`
	TgGroupJoinGame *TgGroup       `gorm:"-"`
	AchivementCode  int            `gorm:"-"`
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

//CheckAchivement is
//every achivement has only 5 levels(0-4)
func (u *User) CheckAchivement() {
	defer u.Update()
	u.achivementCheck(&(u.GamesJoinedAchive), AchivementGamesJoined, u.GamesJoined, 10, 10, true)
	u.achivementCheck(&(u.GamesWonAchive), AchivementGamesWon, u.GamesWon, 0, 5, false)
	u.achivementCheck(&(u.ShootAchive), AchivementShoot, u.ShootCount, 10, 20, true)
	u.achivementCheck(&(u.BetrayAchive), AchivementBetray, u.BetrayCount, 0, 10, false)
	u.achivementCheck(&(u.KillAchive), AchivementKill, u.KillCount, 0, 10, false)
	u.achivementCheck(&(u.GuessKillCountAchive), AchivementGuessKill, u.GuessKillCount, 0, 5, false)
	u.achivementCheck(&(u.SniperKillCountAchive), AchivementSniperKill, u.SniperKillCount, 0, 10, false)
	u.achivementCheck(&(u.KilledByTrapAchive), AchivementKilledByTrap, u.KilledByTrapCount, 0, 5, false)
	u.achivementCheck(&(u.TrapAchive), AchivementTrap, u.TrapCount, 5, 10, true)
	u.achivementCheck(&(u.UnionAchive), AchivementUnion, u.UnionCount, 10, 10, true)
	u.achivementCheck(&(u.UnionSuccessAchive), AchivementUnionSuccess, u.UnionSuccessCount, 0, 10, false)
	u.achivementCheck(&(u.BeUnionedAchive), AchivementBeUnioned, u.BeUnionedCount, 5, 20, true)
}

func (u *User) String() string {
	return fmt.Sprintf("User(TgUserID=%d TgUserName=`%s` Name=`%s` Language=`%s`)",
		u.TgUserID, u.TgUserName, u.Name, u.Language)
}

func (u *User) achivementCheck(achivementLevel *int, achivement, count, base, times int, greaterOrEqual bool) {
	var cond bool
	if base == 0 {
		greaterOrEqual = false
	}
	if greaterOrEqual {
		cond = count >= achiveLevelToCount(base, times, *achivementLevel)
	} else {
		cond = count > achiveLevelToCount(base, times, *achivementLevel)
	}
	if cond && (*achivementLevel) < maxAchivementLevel {
		u.AchivementCode = (*achivementLevel) + 10*achivement
		AchivementRewardedHint <- u
		(*achivementLevel)++
		u.AchiveRewardedCount++
	}
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

func achiveLevelToCount(base int, times int, level int) int {
	var sum = 0
	for i := 1; i <= level; i++ {
		sum += times * i * i * i
	}
	return base + sum
}
