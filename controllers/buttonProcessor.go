package controllers

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"

	"git.wetofu.top/tonychee7000/blackForestBot/lang"

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
		log.Println("DEBUG:", group)
		if int64(act.From.ID) != group.Admin.TgUserID {
			return errors.New("No permission to change language")
		}
		group.Lang = arg
		if err := group.Update(); err != nil {
			return err
		}
		lang.UserLang[act.Message.Chat.ID] = arg

	} else if act.Message.Chat.ID > 0 {
		user, err := models.GetUser(int64(act.Message.Chat.ID))
		if err != nil {
			return err
		}
		user.Language = arg
		if err := user.Update(); err != nil {
			return err
		}
		lang.UserLang[act.Message.Chat.ID] = arg
	}
	DeleteMessageEvent <- tgApi.DeleteMessageConfig{
		ChatID:    act.Message.Chat.ID,
		MessageID: act.Message.MessageID,
	}
	LanguageChangedEvent <- act
	return nil
}

func btnCancelGame(arg string, act *tgApi.CallbackQuery) error {
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	if arg == "" {
		return errors.New("Bad game id")
	}
	ID, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return err
	}

	if err := DelGameQueue(ID,
		models.NewQueueElement(act.Message.Chat.ID, act.Message.MessageID),
	); err != nil {
		return err
	}

	defer func() {
		DeleteMessageEvent <- tgApi.DeleteMessageConfig{
			ChatID:    act.Message.Chat.ID,
			MessageID: act.Message.MessageID,
		}
	}()
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
	OperationApprovedEvent <- act
	return nil
}

func btnUnionReq(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)

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

	// unionreq for none, means skip
	if args[0] == "" {
		RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
			srcPlayer.User.TgUserID, srcPlayer.UnionReq,
			tgApi.InlineKeyboardMarkup{},
		)
		return nil
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return err
	}

	targetPlayer := game.GetPlayer(userID)
	if targetPlayer == nil {
		return errors.New("No such target player")
	}

	// if target has union, you cannot union.
	if targetPlayer.UnionValidation() {
		models.UnionRejectHint <- []*models.Player{targetPlayer, srcPlayer}
	} else {
		models.UnionReqHint <- []*models.Player{srcPlayer, targetPlayer}
	}

	return nil
}

func btnUnionAccept(arg string, act *tgApi.CallbackQuery) error {
	args := strings.SplitN(arg, ",", 2)
	ID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}

	// No one selected.
	if args[0] == "" {
		return nil
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

	if args[0] == "" {
		return nil
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
		player.User.TgUserID, act.Message.MessageID, tgApi.InlineKeyboardMarkup{})
	OperationApprovedEvent <- act
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
	game.AttachOperation(player.SetTrap())
	RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		player.User.TgUserID, act.Message.MessageID, tgApi.InlineKeyboardMarkup{})
	OperationApprovedEvent <- act
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
		player.User.TgUserID, act.Message.MessageID, tgApi.InlineKeyboardMarkup{})
	OperationApprovedEvent <- act
	return nil
}
