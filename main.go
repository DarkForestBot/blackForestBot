package main

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/controllers"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	defer database.DB.Close()
	defer database.Redis.Close()
	bot := bot.NewBot()
	if err := bot.Connect(config.DefaultConfig); err != nil {
		log.Fatalln("FATAL:", err)
	}
	log.Printf("Bot authoirzed by name: %s(%d)", bot.Name(), bot.ID())
	bot.RegisterProcessor(controllers.MessageProcessor)

	go controllers.JoinContoller(bot)
	go controllers.GameController(bot)
	bot.Run()
}
