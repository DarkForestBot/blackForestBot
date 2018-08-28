package controllers

import (
	"errors"
	"strings"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type eventFunc func(*tgApi.Message, ...interface{}) error
type inlineQueryFunc func(string, *tgApi.CallbackQuery, *bot.Bot) error

var commandList map[string]eventFunc
var inlineQueryList map[string]inlineQueryFunc

func init() {
	commandList = make(map[string]eventFunc)
	commandList["start"] = onStart
	commandList["help"] = onHelp
	commandList["startgame"] = onStartGame
	commandList["admin"] = onAdmin
	commandList["extend"] = onExtend
	commandList["players"] = onPlayers
	commandList["flee"] = onFlee
	commandList["setlang"] = onSetLang
	commandList["stats"] = onStat
	commandList["forcestart"] = onForceStart
	commandList["nextgame"] = onNextGame
	inlineQueryList = make(map[string]inlineQueryFunc)
	inlineQueryList["setlang"] = btnSetLang
	inlineQueryList["cancelgame"] = btnCancelGame

}

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
	return fn(arg, act, bot)
}
