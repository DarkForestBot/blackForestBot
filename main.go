package main

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/utils"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("BlackForest running...")
}

func main() {
	defer database.DB.Close()
	defer database.Redis.Close()
	utils.DummyForLoad = 0
	bot.DefaultBot.Run()
}
