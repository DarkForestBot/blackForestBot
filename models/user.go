package models

// User is used in database
type User struct {
	ID                   int
	TgUserID             int    `gorm:"unique;not null;size:50"`
	TgUserName           string `gorm:"not null;size:255"`
	Name                 string `gorm:"not null"`
	GamesJoined          int    `gorm:"default:0;not null"`
	GamesWon             int    `gorm:"default:0;not null"`
	Language             string `gorm:"default:\"English\""`
	ShootCount           int    `gorm:"default:0;not null"`
	BetrayCount          int    `gorm:"default:0;not null"`
	KillCount            int    `gorm:"default:0;not null"`
	UnionCount           int    `gorm:"default:0;not null"`
	UnionSuccessCount    int    `gorm:"default:0;not null"`
	BeUnionedCount       int    `gorm:"default:0;not null"`
	ArchiveRewardedCount int    `gorm:"default:0;not null"`
}
