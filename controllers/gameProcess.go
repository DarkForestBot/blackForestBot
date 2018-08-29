package controllers

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func unionHint(game *models.Game, bot *bot.Bot) {
	for _, player := range game.Players {
		if player.Live == models.PlayerLive &&
			(player.Unioned == nil || player.Unioned.Live == models.PlayerDead) {
			langSet := getLang(player.User.TgUserID)
			msg := tgApi.NewMessage(player.User.TgUserID, lang.T(langSet, "unionhint", nil))
			msg.ParseMode = tgApi.ModeMarkdown
			var btns [][]tgApi.InlineKeyboardButton = make([][]tgApi.InlineKeyboardButton, 0)
			for _, playerB := range game.Players {
				if playerB != player && playerB.Live == models.PlayerLive &&
					(playerB.Unioned == nil || playerB.Unioned.Live == models.PlayerDead) {
					btns = append(
						btns,
						[]tgApi.InlineKeyboardButton{
							tgApi.NewInlineKeyboardButtonData(
								player.User.Name, fmt.Sprintf("unionreq=%d", player.User.TgUserID),
							),
						},
					)
				}
			}
			btns = append(btns,
				[]tgApi.InlineKeyboardButton{
					tgApi.NewInlineKeyboardButtonData(
						lang.T(langSet, "skip", nil), "unionreq=",
					),
				},
			)
			msg.ReplyMarkup = btns
			nmsg, err := bot.Send(msg)
			if err != nil {
				log.Println("ERROR:", err)
			}
			game.MsgSent.UnionOperMsg = append(game.MsgSent.UnionOperMsg, nmsg.MessageID)
		}
	}
}

func sendUnionRequest(from, to *models.Player, bot *bot.Bot) error {
	langSet := getLang(to.User.TgUserID)
	msg := tgApi.NewMessage(to.User.TgUserID, lang.T(langSet, "unionreq", from.User))
	msg.ParseMode = tgApi.ModeMarkdown
	btnAccept := tgApi.NewInlineKeyboardButtonData(
		lang.T(langSet, "accept", nil),
		fmt.Sprintf("unionaccept=%d", from.User.TgUserID),
	)
	btnReject := tgApi.NewInlineKeyboardButtonData(
		lang.T(langSet, "reject", nil),
		"unionaccept=",
	)
	msg.ReplyMarkup = tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(btnAccept, btnReject),
	)
	nmsg, err := bot.Send(msg)
	if err != nil {
		return err
	}
	to.UnionReqs = append(to.UnionReqs, nmsg.MessageID)
	return nil
}
