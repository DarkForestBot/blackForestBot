package bot

import (
	"errors"
	"log"
	"strings"

	"git.wetofu.top/tonychee7000/blackForestBot/controllers"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func messageProcessor(update tgApi.Update, bot *Bot) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	msg := update.Message
	if msg != nil {
		if err := checkJoinEvent(msg, bot); err != nil {
			log.Println("ERROR:", err)
			return
		}
		if err := checkLeaveEvent(msg, bot); err != nil {
			log.Println("ERROR:", err)
			return
		}
		if err := checkAnimationEvent(msg); err != nil {
			log.Println("ERROR:", err)
			return
		}
		if err := checkCommandEvent(msg); err != nil {
			log.Println("ERROR:", err)
			return
		}
	}
	act := update.CallbackQuery
	if act != nil {
		if err := checkCallbackEvent(act, bot); err != nil {
			log.Println("ERROR:", err)
			return
		}
	}
}

func checkJoinEvent(msg *tgApi.Message, bot *Bot) error {
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

	//I don't hear radio!
	if msg.Chat.Type == "channel" {
		if _, err := bot.LeaveChat(tgApi.ChatConfig{
			ChatID: msg.Chat.ID,
		}); err != nil {
			return err
		}
	}

	return controllers.OnJoinAChat(msg)
}

func checkLeaveEvent(msg *tgApi.Message, bot *Bot) error {
	if msg.LeftChatMember != nil && msg.LeftChatMember.ID == bot.ID() {
		return controllers.OnKickoutAChat(msg)
	}
	return nil
}

func checkCommandEvent(msg *tgApi.Message) error {
	if msg.IsCommand() {
		cmd := msg.Command()
		args := msg.CommandArguments()
		fn, ok := controllers.CommandList[cmd]
		if !ok {
			return errors.New("No such command")
		}
		return fn(msg, args)
	}
	return nil
}

func checkAnimationEvent(msg *tgApi.Message) error {
	if msg.Document != nil {
		return controllers.OnReceiveAnimation(msg)
	}
	return nil
}

func checkCallbackEvent(act *tgApi.CallbackQuery, bot *Bot) error {
	a := strings.SplitN(act.Data, "=", 2)
	cmd := a[0]
	arg := a[1]
	fn, ok := controllers.InlineQueryList[cmd]
	if !ok {
		return errors.New("No such inline command")
	}
	if err := fn(arg, act); err != nil {
		return err
	}
	if _, err := bot.AnswerCallbackQuery(
		tgApi.CallbackConfig{CallbackQueryID: act.ID}); err != nil {
		return err
	}
	return nil
}
