package main

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("BlackForest running...")
}

func main() {
	defer database.DB.Close()
	defer database.Redis.Close()

	bot.DefaultBot.Run()
}
