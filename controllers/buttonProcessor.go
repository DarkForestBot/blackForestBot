package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type inlineQueryFunc func(string, *tgApi.CallbackQuery) error

//InlineQueryList is
var InlineQueryList map[string]inlineQueryFunc

func init() {
	InlineQueryList = make(map[string]inlineQueryFunc)
	InlineQueryList["setlang"] = btnSetLang
	InlineQueryList["cancelgame"] = btnCancelGame
	InlineQueryList["x"] = btnShootOne // data: "x=po,groupid"
	InlineQueryList["y"] = btnShootTwo
	InlineQueryList["unionreq"] = btnUnionReq
	InlineQueryList["unionaccept"] = btnUnionAccept
	InlineQueryList["unionreject"] = btnUnionReject
	InlineQueryList["betray"] = btnBetray
	InlineQueryList["trap"] = btnTrap
	InlineQueryList["abort"] = btnAbort
}

func btnSetLang(arg string, act *tgApi.CallbackQuery) error {
	if act.Message.Chat.ID < 0 {
		group, err := models.GetTgGroup(act.Message.Chat.ID)
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

	} else if act.Message.Chat.ID > 0 {
		user, err := models.GetUser(int64(act.Message.From.ID))
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
	}
	DeleteMessageEvent <- tgApi.DeleteMessageConfig{
		ChatID:    act.Message.Chat.ID,
		MessageID: act.Message.MessageID,
	}
	LanguageChangedEvent <- act
	return nil
}

func btnCancelGame(arg string, act *tgApi.CallbackQuery) error {
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
	DeleteMessageEvent <- tgApi.DeleteMessageConfig{
		ChatID:    act.Message.Chat.ID,
		MessageID: act.Message.MessageID,
	}
	return nil
}

func btnShootOne(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}

	x, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}
	player := game.GetPlayer(int64(act.From.ID))
	if player == nil {
		return errors.New("No such player found")
	}

	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	player.ShootX = x
	models.ShootXHint <- player
	return nil
}

func btnShootTwo(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}

	y, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}
	player := game.GetPlayer(int64(act.From.ID))
	if player == nil {
		return errors.New("No such player found")
	}

	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	player.ShootY = y
	pos := game.GetPosition(player.ShootX, player.ShootY)
	if pos == nil {
		return errors.New("Shoot outta map")
	}
	game.AttachOperation(player.Shoot(false, pos))

	models.ShootYHint <- player
	return nil
}

func btnUnionReq(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return err
	}

	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}

	srcPlayer := game.GetPlayer(int64(act.From.ID))
	if srcPlayer == nil {
		return errors.New("No such source player")
	}

	targetPlayer := game.GetPlayer(userID)
	if targetPlayer == nil {
		return errors.New("No such target player")
	}

	models.UnionReqHint <- []*models.Player{srcPlayer, targetPlayer}

	return nil
}

func btnUnionAccept(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return err
	}

	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}

	srcPlayer := game.GetPlayer(int64(act.From.ID))
	if srcPlayer == nil {
		return errors.New("No such source player")
	}

	targetPlayer := game.GetPlayer(userID)
	if targetPlayer == nil {
		return errors.New("No such target player")
	}
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	targetPlayer.Union(srcPlayer)
	models.UnionAcceptHint <- []*models.Player{srcPlayer, targetPlayer}
	return nil
}

func btnUnionReject(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return err
	}

	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}

	srcPlayer := game.GetPlayer(int64(act.From.ID))
	if srcPlayer == nil {
		return errors.New("No such source player")
	}

	targetPlayer := game.GetPlayer(userID)
	if targetPlayer == nil {
		return errors.New("No such target player")
	}
	models.UnionRejectHint <- []*models.Player{srcPlayer, targetPlayer}
	return nil
}

func btnBetray(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}
	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}
	player := game.GetPlayer(int64(act.From.ID))
	game.AttachOperation(player.Shoot(true, nil))
	RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		ID, player.OperationMsg, tgApi.InlineKeyboardMarkup{})
	return nil
}

func btnTrap(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}
	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}
	player := game.GetPlayer(int64(act.From.ID))
	player.TrapSet = true
	RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		ID, player.OperationMsg, tgApi.InlineKeyboardMarkup{})
	return nil
}

func btnAbort(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}
	game, ok := gameList[ID]
	if !ok {
		return errors.New("No such game")
	}
	player := game.GetPlayer(int64(act.From.ID))
	game.AttachOperation(player.Abort())
	RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		ID, player.OperationMsg, tgApi.InlineKeyboardMarkup{})
	return nil
}
