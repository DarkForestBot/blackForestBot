package controllers

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func TestProcessor(update tgApi.Update, bot *bot.Bot) error {
	var err error
	msg := update.Message
	if msg != nil {
		btn := tgApi.NewInlineKeyboardButtonData("Test", "k=123")
		row := tgApi.NewInlineKeyboardRow(btn)
		mk := tgApi.NewInlineKeyboardMarkup(row)

		reply := tgApi.NewMessage(msg.Chat.ID, "*?*")
		reply.ReplyMarkup = mk
		reply.ParseMode = tgApi.ModeMarkdown
		reply.ReplyToMessageID = msg.MessageID
		log.Println("MessageRecv:", msg.From.String())
		_, err = bot.Send(reply)
	}
	act := update.CallbackQuery
	if act != nil {
		log.Println(act.Data)
		bot.AnswerCallbackQuery(tgApi.CallbackConfig{
			CallbackQueryID: act.ID,
			URL:             fmt.Sprintf("https://t.me/%s?start=%s", bot.Name(), act.ID),
		})
	}
	return err
}
