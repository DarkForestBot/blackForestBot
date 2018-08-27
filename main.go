package main

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/controllers"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	database.DB.LogMode(config.DefaultConfig.Debug)
	database.DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.User{}, &models.TgGroup{})
	defer database.DB.Close()
	bot := bot.NewBot()
	err := bot.Connect(config.DefaultConfig)
	if err != nil {
		log.Fatalln("FATAL:", err)
	}

	log.Printf("Bot authoirzed by name: %s(%d)", bot.Name(), bot.ID())
	bot.RegisterProcessor(controllers.TestProcessor)
	bot.Run()
}
