package controllers

import (
	"fmt"
	"log"
	"time"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type JoinTimeExtender struct {
	ChatID     int64
	ExtendTime int
}

var gameList map[int64]*models.Game
var ChJoinTimeExtender chan JoinTimeExtender
var ChGameExtender chan *models.Game
var ChGameCanceller chan int64
var ChGameGetter chan int64
var ChGameRecv chan *models.Game

func init() {
	gameList = make(map[int64]*models.Game)
	ChJoinTimeExtender = make(chan JoinTimeExtender)
	ChGameExtender = make(chan *models.Game)
	ChGameCanceller = make(chan int64)
	ChGameGetter = make(chan int64)
	ChGameRecv = make(chan *models.Game)
}

func JoinContoller(
	chJoinTimeExtender chan JoinTimeExtender,
	chGameExtender chan *models.Game,
	chGameCanceller chan int64,
	chGameGetter chan int64,
	chGameRecv chan *models.Game,
	bot *bot.Bot) {
	log.Println("JoinController run.")
	for {
		select {
		case c := <-chJoinTimeExtender:
			langSet := getLang(c.ChatID)
			game := gameList[c.ChatID]
			game.JoinTime += c.ExtendTime
			if game.JoinTime > 300 {
				game.JoinTime = 300
			}
			msg := tgApi.NewMessage(c.ChatID, lang.T(langSet, "jointime",
				fmt.Sprintf("%d:%d", game.JoinTime/60, game.JoinTime%60),
			))
			msg.ReplyMarkup = joinButton(c.ChatID, bot)
			m, err := bot.Send(msg)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
			game.MsgSent.JoinTimeMsg = append(game.MsgSent.JoinTimeMsg, &m)
		case <-time.Tick(time.Second):
			for k, v := range gameList {
				if v == nil {
					continue
				}
				if v.JoinTime > 0 && v.Status == models.GameNotStart {
					v.JoinTime--
				}
				langSet := getLang(k)
				switch v.JoinTime {
				case 60:
					fallthrough
				case 30:
					fallthrough
				case 10:
					msg := tgApi.NewMessage(k, lang.T(langSet, "jointime",
						fmt.Sprintf("%d:%d", v.JoinTime/60, v.JoinTime%60),
					))
					msg.ReplyMarkup = joinButton(k, bot)
					m, err := bot.Send(msg)
					if err != nil {
						log.Println("ERROR:", err)
						continue
					}
					gameList[k].MsgSent.JoinTimeMsg = append(gameList[k].MsgSent.JoinTimeMsg, &m)

				case 0:
					if v.Status != models.GameNotStart {
						continue
					}
					msgSent := gameList[k].MsgSent
					m1 := tgApi.EditMessageReplyMarkupConfig{
						BaseEdit: tgApi.BaseEdit{
							ChatID:      k,
							MessageID:   msgSent.StartMsg.MessageID,
							ReplyMarkup: nil,
						},
					}
					bot.Send(m1)
					for _, n := range msgSent.JoinTimeMsg {
						m := tgApi.NewDeleteMessage(k, n.MessageID)
						bot.DeleteMessage(m)
					}

					err := v.Start()
					if err != nil {
						v = nil
						msg := tgApi.NewMessage(k, lang.T(langSet, "gamecancelled", nil))
						bot.Send(msg)
						log.Println("ERROR:", err)
					}
					msg := tgApi.NewMessage(k, lang.T(langSet, "gamestart", nil))
					bot.Send(msg)
				}
			}
		case <-time.Tick(10 * time.Second):
			for k, v := range gameList {
				langSet := getLang(k)
				update := tgApi.NewEditMessageText(k, v.MsgSent.PlayerList.MessageID, lang.T(langSet, "players", v.Users))
				bot.Send(update)
			}
		}
	}
}
