package controllers

import (
	"errors"
	"strings"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//MessageProcessor is
func MessageProcessor(update tgApi.Update, bot *bot.Bot) error {
	msg := update.Message
	if msg != nil {
		if err := checkJoinEvent(msg, bot, onJoinAChat); err != nil {
			return err
		}
		if err := checkLeaveEvent(msg, bot, onKickoutAChat); err != nil {
			return err
		}
		if err := checkAnimationEvent(msg, bot, onReceiveAnimation); err != nil {
			return err
		}
		if err := checkCommandEvent(msg, bot); err != nil {
			return err
		}
	}
	act := update.CallbackQuery
	if act != nil {
		return inlineQueryProcessor(act, bot)
	}
	return nil
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

	return fn(msg, bot)
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

func checkAnimationEvent(msg *tgApi.Message, bot *bot.Bot, fn eventFunc) error {
	if msg.Document != nil {
		return fn(msg)
	}
	return nil
}

func inlineQueryProcessor(act *tgApi.CallbackQuery, bot *bot.Bot) error {
	a := strings.SplitN(act.Data, "=", 2)
	cmd := a[0]
	arg := a[1]
	fn, ok := inlineQueryList[cmd]
	if !ok {
		return errors.New("No such inline command")
	}
	if err := fn(arg, act, bot); err != nil {
		return err
	}
	if _, err := bot.AnswerCallbackQuery(
		tgApi.CallbackConfig{CallbackQueryID: act.ID}); err != nil {
		return err
	}
	return nil
}
