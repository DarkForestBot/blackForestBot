package controllers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//JoinTimeExtender is
type JoinTimeExtender struct {
	ChatID     int64
	ExtendTime int
}

var gameList map[int64]*models.Game

func init() {
	gameList = make(map[int64]*models.Game)
}

//JoinContoller is
func JoinContoller(bot *bot.Bot) {
	log.Println("JoinController run.")
	for {
		select {
		case <-time.Tick(time.Second):
			for k, v := range gameList {
				if v == nil {
					continue
				}
				var lock sync.Mutex
				lock.Lock()
				if v.TimeLeft > 0 && v.Status == models.GameNotStart {
					v.TimeLeft--
				}
				langSet := getLang(k)
				switch v.TimeLeft {
				case 60:
					fallthrough
				case 30:
					fallthrough
				case 10:
					msg := tgApi.NewMessage(k, lang.T(langSet, "jointime",
						fmt.Sprintf("%d:%d", v.TimeLeft/60, v.TimeLeft%60),
					))
					msg.ReplyMarkup = joinButton(k, bot)
					m, err := bot.Send(msg)
					if err != nil {
						log.Println("ERROR:", err)
						continue
					}
					gameList[k].MsgSent.JoinTimeMsg = append(gameList[k].MsgSent.JoinTimeMsg, m.MessageID)

				case 0:
					if v.Status != models.GameNotStart {
						continue
					}
					msgSent := gameList[k].MsgSent
					m1 := tgApi.EditMessageReplyMarkupConfig{
						BaseEdit: tgApi.BaseEdit{
							ChatID:      k,
							MessageID:   msgSent.StartMsg,
							ReplyMarkup: nil,
						},
					}
					bot.Send(m1)
					for _, n := range msgSent.JoinTimeMsg {
						m := tgApi.NewDeleteMessage(k, n)
						bot.DeleteMessage(m)
					}

					err := v.Start()
					if err != nil {
						log.Println(v, "not start for timed out.")
						v = nil
						msg := tgApi.NewMessage(k, lang.T(langSet, "gamecancelled", nil))
						bot.Send(msg)
						log.Println("ERROR:", err)
						continue
					}
					msg := tgApi.NewMessage(k, lang.T(langSet, "gamestart", nil))
					bot.Send(msg)
				}
				lock.Unlock()
			}
		case <-time.Tick(5 * time.Second):
			for k, v := range gameList {
				langSet := getLang(k)
				update := tgApi.NewEditMessageText(k, v.MsgSent.PlayerList, lang.T(langSet, "players", v.Users))
				bot.Send(update)
			}
		}
	}
}

//GameController is
func GameController(bot *bot.Bot) {
	log.Println("GameController run.")
	for {
		select {
		case <-time.Tick(time.Second):
			var lock sync.Mutex
			lock.Lock()
			for _, game := range gameList {
				if game != nil && game.Status == models.GameStart && game.TimeLeft > 0 {
					game.TimeLeft--
				}
			}
			lock.Unlock()
		default:
			var lock sync.Mutex
			lock.Lock()
			for _, game := range gameList {
				if game != nil {
					if game.Status == models.GameStart {
						livePlayers := make([]*models.Player, 0)
						for _, player := range game.Players {
							if player.Live {
								livePlayers = append(livePlayers, player)
							}
						}
						if len(livePlayers) == 1 {
							//Someone win
						} else if len(livePlayers) == 0 {
							//All dead
						}
						if game.IsDay == models.GameIsDay {
							if game.TimeLeft <= 0 {
								game.IsDay = models.GameIsNight
								game.HintSent = true
								game.TimeLeft = consts.OneMinute
								//disable all day operation
							}
							if !game.HintSent {
								//send playerlist and union hint
								unionHint(game, bot)
								game.HintSent = true
							}
						} else {
							if game.TimeLeft <= 0 || len(game.Operations) == len(livePlayers) {
								gameLogic(game)
								//send result
								game.IsDay = models.GameIsDay
								game.TimeLeft = consts.TwoMinutes
								game.HintSent = false
								game.Round++
							}
							if !game.HintSent {
								game.HintSent = true
							}
							//send night hint
						}
					} else if game.Status == models.GameFinished {
						game = nil
					}
				}
			}
			lock.Unlock()
		}
	}
}
