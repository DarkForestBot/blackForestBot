package controllers

import (
	"errors"
	"fmt"
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type eventFunc func(*tgApi.Message, ...interface{}) error

var commandList map[string]eventFunc

func init() {
	commandList = make(map[string]eventFunc)
	commandList["start"] = onStart
	commandList["help"] = onHelp
}

func MessageProcessor(update tgApi.Update, bot *bot.Bot) error {
	var err error
	msg := update.Message
	if msg != nil {
		err = checkJoinEvent(msg, bot, onJoinAChat)
		if err != nil {
			return err
		}
		err = checkLeaveEvent(msg, bot, onKickoutAChat)
		if err != nil {
			return err
		}
		err = checkCommandEvent(msg, bot)
		if err != nil {
			return err
		}
	}
	act := update.CallbackQuery
	if act != nil {
		log.Println(act.Data)
		bot.AnswerCallbackQuery(tgApi.CallbackConfig{
			CallbackQueryID: fmt.Sprintf("ID-%s", act.ID),
			Text:            "Clicked.",
		})
	}
	return err
}

func checkJoinEvent(msg *tgApi.Message, bot *bot.Bot, fn eventFunc) error {
	if msg.NewChatMembers != nil {
		members := *(msg.NewChatMembers)
		for _, m := range members {
			if m.ID == bot.ID() {
				break
			}
		}
	} else {
		return nil
	}

	return fn(msg)
}

func checkLeaveEvent(msg *tgApi.Message, bot *bot.Bot, fn eventFunc) error {
	if msg.LeftChatMember != nil && msg.LeftChatMember.ID == bot.ID() {
		return fn(msg)
	}
	return nil
}

func checkCommandEvent(msg *tgApi.Message, bot *bot.Bot) error {
	if msg.IsCommand() {
		cmd := msg.Command()
		args := msg.CommandArguments()
		fn, ok := commandList[cmd]
		if !ok {
			return errors.New("No such command")
		}
		return fn(msg, bot, args)
	}
	return nil
}
