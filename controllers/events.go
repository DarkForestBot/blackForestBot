package controllers

import (
	"errors"
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func onJoinAChat(msg *tgApi.Message, other ...interface{}) error {
	if msg.Chat.ID > 0 {
		return errors.New("Bad tg group")
	}
	group := new(models.TgGroup)
	if err := database.DB.Where(models.TgGroup{TgGroupID: msg.Chat.ID}).Assign(
		models.TgGroup{
			Name: msg.Chat.Title,
			Admin: models.User{
				TgUserID:   msg.From.ID,
				Name:       fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName),
				TgUserName: msg.From.UserName,
			},
		},
	).FirstOrCreate(group).Error; err != nil {
		return err
	}

	log.Printf("Group `%s` registered.\n", group.Name)
	return nil
}

func onKickoutAChat(msg *tgApi.Message, other ...interface{}) error {
	if msg.Chat.ID > 0 {
		return errors.New("Bad tg group")
	}
	log.Printf("`%s %s(%d)` kicks you out from `%s(%d)`", msg.From.FirstName, msg.From.LastName, msg.From.ID, msg.Chat.Title, msg.Chat.ID)
	return nil
}

func onStart(msg *tgApi.Message, other ...interface{}) error {
	return nil
}

func onHelp(msg *tgApi.Message, other ...interface{}) error {
	var (
		Bot  *bot.Bot
		args string
		ok   bool
	)
	for _, o := range other {
		if Bot == nil {
			if Bot, ok = o.(*bot.Bot); ok {
				continue
			}
		}
		if args == "" {
			if args, ok = o.(string); ok {
				continue
			}
		}
	}
	if Bot == nil {
		return errors.New("no bot found")
	}
	reply := tgApi.NewMessage(msg.Chat.ID, fmt.Sprintf("*test* %s", args))
	reply.ReplyToMessageID = msg.MessageID
	reply.ParseMode = tgApi.ModeMarkdown
	_, err := Bot.Send(reply)
	return err
}
