package bot

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/controllers"

	"git.wetofu.top/tonychee7000/blackForestBot/models"

	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var threadLimitPool chan int

func init() {
	threadLimitPool = make(chan int, config.DefaultConfig.ThreadLimit)
}

func releaseThreadPool() {
	<-threadLimitPool
}

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

func makeUnionButtons(game *models.Game, exceptPlayer *models.Player) tgApi.InlineKeyboardMarkup {
	var btns tgApi.InlineKeyboardMarkup
	btns.InlineKeyboard = make([][]tgApi.InlineKeyboardButton, 0)

	for _, player := range game.Players {
		if player.Live && !player.UnionValidation() && player != exceptPlayer {
			btns.InlineKeyboard = append(
				btns.InlineKeyboard,
				[]tgApi.InlineKeyboardButton{
					tgApi.NewInlineKeyboardButtonData(
						player.User.Name,
						fmt.Sprintf("unionreq=%d,%d", player.User.TgUserID, game.TgGroup.TgGroupID),
					),
				},
			)
		}
	}
	if len(btns.InlineKeyboard) != 0 {
		btns.InlineKeyboard = append(
			btns.InlineKeyboard,
			[]tgApi.InlineKeyboardButton{
				tgApi.NewInlineKeyboardButtonData(
					lang.T(
						exceptPlayer.User.Language, "skip", nil,
					),
					fmt.Sprintf("unionreq=,%d", game.TgGroup.TgGroupID),
				),
			},
		)
	}
	return btns
}

func makeNightOperations(TgGroupID int64, player *models.Player, step int) tgApi.InlineKeyboardMarkup {
	var btns tgApi.InlineKeyboardMarkup
	btns.InlineKeyboard = make([][]tgApi.InlineKeyboardButton, 0)
	if player.UnionValidation() {
		btns.InlineKeyboard = append(
			btns.InlineKeyboard,
			[]tgApi.InlineKeyboardButton{
				tgApi.NewInlineKeyboardButtonData(
					lang.T(player.User.Language, "betray", nil),
					fmt.Sprintf("betray=,%d", TgGroupID),
				),
			},
		)
		btns.InlineKeyboard = append(
			btns.InlineKeyboard,
			[]tgApi.InlineKeyboardButton{
				tgApi.NewInlineKeyboardButtonData(
					lang.T(player.User.Language, "trap", nil),
					fmt.Sprintf("trap=,%d", TgGroupID),
				),
			},
		)
	}
	var br = make([]tgApi.InlineKeyboardButton, 0)
	for i := 0; i < len(game.Players)*2; i++ {
		var data string
		if step == 0 {
			if i == player.Position.X {
				continue
			}
			data = fmt.Sprintf("x=%d,%d", i, TgGroupID)
		} else if step == 1 {
			if i == player.Position.Y {
				continue
			}
			data = fmt.Sprintf("y=%d,%d", i, TgGroupID)
		}
		br = append(br,
			tgApi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d", i), data,
			),
		)
	}
	for i := 0; i < len(br); i += 5 {
		if len(br)-i <= len(br)%2 {
			btns.InlineKeyboard = append(
				btns.InlineKeyboard,
				br[i:],
			)
		} else {
			btns.InlineKeyboard = append(
				btns.InlineKeyboard,
				br[i:5],
			)
		}
	}
	btns.InlineKeyboard = append(
		btns.InlineKeyboard,
		[]tgApi.InlineKeyboardButton{
			tgApi.NewInlineKeyboardButtonData(
				lang.T(player.User.Language, "abort", nil),
				fmt.Sprintf("abort=,%d", TgGroupID),
			),
		},
	)

	return btns
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
