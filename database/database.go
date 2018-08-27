package database

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //MySQL driver
)

var DB *gorm.DB

func init() {
	var err error
	DSN := config.DefaultConfig.Database
	DB, err = gorm.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}
	log.Println("Database connected.")
}
