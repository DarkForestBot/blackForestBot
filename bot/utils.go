package bot

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/controllers"

	"git.wetofu.top/tonychee7000/blackForestBot/models"

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

func (b *Bot) startGameClearMessage(game *models.Game) {
	//Step I: delete join time left message
	for _, id := range game.MsgSent.JoinTimeMsg {
		controllers.DeleteMessageEvent <- tgApi.DeleteMessageConfig{
			ChatID:    game.TgGroup.TgGroupID,
			MessageID: id,
		}
	}
	//Step II: remove join button
	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		game.TgGroup.TgGroupID, game.MsgSent.StartMsg,
		tgApi.InlineKeyboardMarkup{},
	)
}
