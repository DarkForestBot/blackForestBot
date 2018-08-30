package controllers

import (
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messageUtils(args []interface{}) (*bot.Bot, string) {
	var (
		Bot     *bot.Bot
		argText string
		ok      bool
	)
	for _, o := range args {
		if Bot == nil {
			if Bot, ok = o.(*bot.Bot); ok {
				continue
			}
		}
		if argText == "" {
			if argText, ok = o.(string); ok {
				continue
			}
		}
	}
	return Bot, argText
}

func getLang(ID int64) string {
	var str string
	if err := database.Redis.Get(fmt.Sprintf(consts.LangSetFormatString, ID)).Scan(&str); err != nil {
		log.Printf("WARNING: error %v", err)
		return "English"
	}
	return str
}

func getUser(ID int64) (*models.User, error) {
	return models.GetUser(ID)
}

func getTgGroup(ID int64) (*models.TgGroup, error) {
	return models.GetTgGroup(ID)
}

func markdownReply(ID int64, key string, msg *tgApi.Message, bot *bot.Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(ID)
	reply := tgApi.NewMessage(msg.Chat.ID, lang.T(langSet, key, other))
	reply.ReplyToMessageID = msg.MessageID
	reply.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func markdownMessage(ID int64, key string, bot *bot.Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(ID)
	reply := tgApi.NewMessage(ID, lang.T(langSet, key, other))
	reply.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func gifMessageReply(ID int64, key string, image string, msg *tgApi.Message, bot *bot.Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(ID)
	reply := tgApi.NewDocumentShare(ID, image)
	reply.Caption = lang.T(langSet, key, other)
	reply.MimeType = "video/mp4"
	reply.ParseMode = tgApi.ModeMarkdown
	reply.ReplyToMessageID = msg.MessageID
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func gifMessage(ID int64, key string, image string, bot *bot.Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(ID)
	reply := tgApi.NewDocumentShare(ID, image)
	reply.Caption = lang.T(langSet, key, other)
	reply.MimeType = "video/mp4"
	reply.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func joinButton(ID int64, bot *bot.Bot) tgApi.InlineKeyboardMarkup {
	langSet := getLang(ID)
	joinButton := tgApi.NewInlineKeyboardButtonURL(
		lang.T(langSet, "join", nil),
		fmt.Sprintf("https://t.me/%s?start=%d", bot.Name(), ID),
	)
	return tgApi.NewInlineKeyboardMarkup(tgApi.NewInlineKeyboardRow(joinButton))
}

func startGamePM(msg *tgApi.Message, bot *bot.Bot) error {
	var gameQueue []int64
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueFormatString, msg.Chat.ID),
	).Scan(&gameQueue); err != nil {
		return err
	}
	for _, i := range gameQueue {
		markdownMessage(i, "newgame", bot, msg.Chat.Title)
	}
	return nil
}
