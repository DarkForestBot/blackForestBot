package controllers

import (
	"errors"
	"fmt"
	"strconv"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type inlineQueryFunc func(string, *tgApi.CallbackQuery, *bot.Bot) error

var inlineQueryList map[string]inlineQueryFunc

func init() {
	inlineQueryList = make(map[string]inlineQueryFunc)
	inlineQueryList["setlang"] = btnSetLang
	inlineQueryList["cancelgame"] = btnCancelGame
}

func btnSetLang(arg string, act *tgApi.CallbackQuery, bot *bot.Bot) error {
	if act.Message.Chat.ID < 0 {
		group, err := getTgGroup(act.Message.Chat.ID)
		if err != nil {
			return err
		}
		if int64(act.Message.From.ID) != group.Admin.TgUserID {
			return errors.New("No permission to change language")
		}
		group.Lang = arg
		if err := group.Update(); err != nil {
			return err
		}
		if err := database.Redis.Set(
			fmt.Sprintf(consts.LangSetFormatString, act.Message.Chat.ID),
			arg, -1,
		).Err(); err != nil {
			return err
		}
		if _, err := bot.DeleteMessage(tgApi.DeleteMessageConfig{
			ChatID:    act.Message.Chat.ID,
			MessageID: act.Message.MessageID,
		}); err != nil {
			return err
		}
	} else if act.Message.Chat.ID > 0 {
		user, err := getUser(int64(act.Message.From.ID))
		if err != nil {
			return err
		}
		user.Language = arg
		if err := user.Update(); err != nil {
			return err
		}
		if err := database.Redis.Set(
			fmt.Sprintf(consts.LangSetFormatString, act.Message.From.ID),
			arg, -1,
		).Err(); err != nil {
			return err
		}
		if _, err := bot.DeleteMessage(tgApi.DeleteMessageConfig{
			ChatID:    act.Message.Chat.ID,
			MessageID: act.Message.MessageID,
		}); err != nil {
			return err
		}
	}
	if _, err := markdownMessage(act.Message.Chat.ID, "langchanged", bot, nil); err != nil {
		return err
	}
	return nil
}

func btnCancelGame(arg string, act *tgApi.CallbackQuery, bot *bot.Bot) error {
	ID, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return err
	}
	var gameQueue []int64
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueFormatString, ID),
	).Scan(&gameQueue); err != nil {
		return err
	}
	for i, k := range gameQueue {
		if k == int64(act.From.ID) {
			gameQueue = append(gameQueue[:i], gameQueue[i+1:]...)
		}
	}
	if err := database.Redis.Set(
		fmt.Sprintf(consts.GameQueueFormatString, ID),
		gameQueue, -1,
	).Err(); err != nil {
		return err
	}
	if _, err := bot.DeleteMessage(tgApi.DeleteMessageConfig{
		ChatID:    act.Message.Chat.ID,
		MessageID: act.Message.MessageID,
	}); err != nil {
		return err
	}
	return nil
}
