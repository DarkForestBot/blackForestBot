package database

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //MySQL driver
)

var DB *gorm.DB
var Redis *redis.Client

func init() {
	var err error
	DSN := config.DefaultConfig.Database
	DB, err = gorm.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}
	log.Println("Database connected.")
	DB.LogMode(config.DefaultConfig.Debug)
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.User{}, &models.TgGroup{})
	opt, err := config.DefaultConfig.RedisConfig()
	if err != nil {
		panic(err)
	}
	Redis = redis.NewClient(opt)
}
