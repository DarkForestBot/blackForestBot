package bot

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/controllers"

	"git.wetofu.top/tonychee7000/blackForestBot/models"

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
	str, ok := lang.UserLang[ID]
	if ok {
		return str
	}
	log.Printf("No language set found by `%d`, use default `English`", ID)
	return "English"
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

func makeNightOperations(TgGroupID int64, player *models.Player, playerCount int, step int) tgApi.InlineKeyboardMarkup {
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
	for i := 0; i < playerCount*2; i++ {
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

	for i := 0; i < len(br); i += 5 { // except myself.
		if len(br)-i <= len(br)%5 {
			btns.InlineKeyboard = append(
				btns.InlineKeyboard,
				br[i:],
			)
		} else {
			btns.InlineKeyboard = append(
				btns.InlineKeyboard,
				br[i:i+5],
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

func (b *Bot) makeReplay(game *models.Game) error {
	langSet := getLang(game.TgGroup.TgGroupID)
	var report = "#Replay\n"
	for i, ops := range game.GlobalOperations {
		var finish = false
		report += lang.T(langSet, "replay_round", i+1)
		for _, op := range ops {
			/*
				if op.IsResult {
					var none = true
					if op.Killed != "" {
						none = false
						if op.Action == models.Betray {
							report += lang.T(langSet, "replay_betrayed", op)
						} else {
							report += lang.T(langSet, "replay_killed", op)
						}
					}
					if op.BeKilled {
						none = false
						report += lang.T(langSet, "replay_bekilled", op)
					}
					if op.BeBeast {
						none = false
						report += lang.T(langSet, "replay_bebeast", op)
					}
					if op.Survive {
						none = false
						report += lang.T(langSet, "replay_survive", op)
					}
					if none {
						if op.Player.User != nil {
							report += lang.T(langSet, "replay_none", op)
						}
					}

				} else {*/
			if op.Finally {
				report += "\n"
				if op.Player.User == nil {
					report += lang.T(langSet, "replay_lose", nil)
				} else {
					report += lang.T(langSet, "replay_win", op)
				}
				report += "\n"
				finish = true
				break
			}

			switch op.Action {
			case models.Shoot:
				if op.Target != nil {
					report += lang.T(langSet, "replay_shoot", op)
				}
			case models.Abort:
				report += lang.T(langSet, "replay_abort", op)
			case models.Trap:
				report += lang.T(langSet, "replay_trap", op)
			}
			for _, r := range op.Result {
				if r.Killed != "" {
					if op.Action == models.Betray {
						report += lang.T(langSet, "replay_betrayed", r)
					} else {
						report += lang.T(langSet, "replay_killed", r)
					}
				}
				if r.BeKilled {
					report += lang.T(langSet, "replay_bekilled", r)
				}
				if r.BeBeast {
					report += lang.T(langSet, "replay_bebeast", r)
				}
				if r.Survive {
					report += lang.T(langSet, "replay_survive", r)
				}
				if r.None {
					if r.Who != nil {
						report += lang.T(langSet, "replay_none", op)
					}
				}
			}
			report += "\n"
		}

		if finish {
			break
		}

		if len(report) > 2048 {
			msg := tgApi.NewMessage(game.TgGroup.TgGroupID, report)
			msg.ParseMode = tgApi.ModeMarkdown
			if _, err := b.Send(msg); err != nil {
				return err
			}
			report = "#Replay\n"
		} else {
			report += "--------\n"
		}
	}
	msg := tgApi.NewMessage(game.TgGroup.TgGroupID, report)
	msg.ParseMode = tgApi.ModeMarkdown
	if _, err := b.Send(msg); err != nil {
		return err
	}
	return nil
}
