package bot

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getLang(ID int64) string {
	var str string
	if err := database.Redis.Get(fmt.Sprintf(consts.LangSetFormatString, ID)).Scan(&str); err != nil {
		log.Printf("WARNING: error %v", err)
		return "English"
	}
	return str
}

func joinButton(ID int64, bot *Bot) tgApi.InlineKeyboardMarkup {
	langSet := getLang(ID)
	joinButton := tgApi.NewInlineKeyboardButtonURL(
		lang.T(langSet, "join", nil),
		fmt.Sprintf("https://t.me/%s?start=%d", bot.Name(), ID),
	)
	return tgApi.NewInlineKeyboardMarkup(tgApi.NewInlineKeyboardRow(joinButton))
}

func startGamePM(msg *tgApi.Message) error {
	var gameQueue []int64
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueFormatString, msg.Chat.ID),
	).Scan(&gameQueue); err != nil {
		return err
	}
	for _, i := range gameQueue {
		markdownMessage(i, "newgame", msg.Chat.Title)
	}
	return nil
}
