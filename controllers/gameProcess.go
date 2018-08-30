package controllers

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func winHint(game *models.Game, winner *models.Player, bot *bot.Bot) error {
	winner.User.GamesWon++
	winner.User.Update()
	if _, err := gifMessage(game.TgGroup.TgGroupID, "wine", config.DefaultImages.Win, bot, winner.User); err != nil {
		return err
	}
	return nil
}

func loseHint(game *models.Game, bot *bot.Bot) error {
	if _, err := gifMessage(game.TgGroup.TgGroupID, "lose", config.DefaultImages.Lose, bot, nil); err != nil {
		return err
	}
	return nil
}

func unionHint(game *models.Game, bot *bot.Bot) {
	for _, player := range game.Players {
		if player.Live == models.PlayerLive &&
			(player.Unioned == nil || player.Unioned.Live == models.PlayerDead) {
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
			langSet := getLang(player.User.TgUserID)
			btns = append(btns,
				[]tgApi.InlineKeyboardButton{
					tgApi.NewInlineKeyboardButtonData(
						lang.T(langSet, "skip", nil), "unionreq=",
					),
				},
			)
			nmsg, err := markdownMessage(
				player.User.TgUserID, "unionhint", bot, nil,
				tgApi.InlineKeyboardMarkup{
					InlineKeyboard: btns,
				},
			)
			if err != nil {
				log.Println("ERROR:", err)
			}
			game.MsgSent.UnionOperMsg = append(game.MsgSent.UnionOperMsg, nmsg.MessageID)
		}
	}
}

func sendUnionRequest(from, to *models.Player, bot *bot.Bot) error {
	langSet := getLang(to.User.TgUserID)
	btnAccept := tgApi.NewInlineKeyboardButtonData(
		lang.T(langSet, "accept", nil),
		fmt.Sprintf("unionaccept=%d", from.User.TgUserID),
	)
	btnReject := tgApi.NewInlineKeyboardButtonData(
		lang.T(langSet, "reject", nil),
		"unionaccept=",
	)
	nmsg, err := markdownMessage(
		to.User.TgUserID, "unionreq", bot, from.User,
		tgApi.NewInlineKeyboardMarkup(
			tgApi.NewInlineKeyboardRow(btnAccept, btnReject),
		),
	)
	if err != nil {
		return err
	}
	to.UnionReqs = append(to.UnionReqs, nmsg.MessageID)
	return nil
}
