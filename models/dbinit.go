package models

import "git.wetofu.top/tonychee7000/blackForestBot/database"

func init() {
	database.DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&User{}, &TgGroup{})
}
