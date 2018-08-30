package controllers

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"git.wetofu.top/tonychee7000/blackForestBot/basis"
	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type eventFunc func(*tgApi.Message, ...interface{}) error

var isAdminMode bool
var commandList map[string]eventFunc

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
	isAdminMode = false
}

func onJoinAChat(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot == nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID > 0 {
		return errors.New("Bad tg group")
	}
	// No channel allowed.
	if msg.Chat.Type == "channel" {
		_, err := bot.LeaveChat(tgApi.ChatConfig{
			ChatID: msg.Chat.ID,
		})
		return err
	}
	group := new(models.TgGroup)
	if err := database.DB.Where(models.TgGroup{TgGroupID: msg.Chat.ID}).Assign(
		models.TgGroup{
			Name: msg.Chat.Title,
			Admin: models.User{
				TgUserID:   int64(msg.From.ID),
				Name:       fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName),
				TgUserName: msg.From.UserName,
			},
			Active: true,
		},
	).FirstOrCreate(group).Error; err != nil {
		return err
	}
	if err := database.Redis.Set(
		fmt.Sprintf(consts.LangSetFormatString, group.TgGroupID),
		group.Lang, -1,
	).Err(); err != nil {
		return err
	}
	log.Printf("Group `%s` registered.\n", group.Name)
	return nil
}

func onKickoutAChat(msg *tgApi.Message, other ...interface{}) error {
	if msg.Chat.ID > 0 {
		return errors.New("Bad tg group")
	}
	group := new(models.TgGroup)
	if err := database.DB.Where(models.TgGroup{TgGroupID: msg.Chat.ID}).First(group).Error; err != nil {
		return err
	}
	if group.ID != 0 {
		group.Active = false
		if err := group.Update(); err != nil {
			return err
		}
	}
	log.Printf("`%s %s(%d)` kicks you out from `%s(%d)`",
		msg.From.FirstName, msg.From.LastName, msg.From.ID, msg.Chat.Title, msg.Chat.ID)
	return nil
}

func onStart(msg *tgApi.Message, other ...interface{}) error {
	bot, args := messageUtils(other)
	if bot == nil {
		return errors.New("no bot found")
	}
	user := new(models.User)
	if err := database.DB.Where(models.User{TgUserID: int64(msg.From.ID)}).Assign(
		models.User{
			TgUserName: msg.From.UserName,
			Name:       fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName),
		},
	).FirstOrCreate(user).Error; err != nil {
		return err
	}
	if err := database.Redis.Set(
		fmt.Sprintf(consts.LangSetFormatString, user.TgUserID),
		user.Language, -1,
	).Err(); err != nil {
		return err
	}
	log.Printf("User `%s(%d)` registered.\n", user.Name, user.TgUserID)

	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	if args != "" {
		id, err := strconv.ParseInt(args, 10, 64)
		if err != nil {
			return err
		}
		game := gameList[id]
		if game != nil && game.Status == models.GameNotStart {
			game.Join(user)
			if _, err := markdownMessage(game.TgGroup.TgGroupID, "joingame", bot, user); err != nil {
				return err
			}
		}
	}

	return nil
}

func onStartGame(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot == nil {
		return errors.New("no bot found")
	}
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	if msg.Chat.ID < 0 {
		user, err := getUser(int64(msg.From.ID))
		if err != nil {
			return err
		}
		group, err := getTgGroup(msg.Chat.ID)
		if err != nil {
			return err
		}

		game := gameList[msg.Chat.ID]
		if game != nil && game.TgGroup.TgGroupID == msg.Chat.ID {
			if _, err := markdownReply(msg.Chat.ID, "hasgame", msg, bot, nil); err != nil {
				return err
			}
		}

		game = models.NewGame(group)
		gameList[msg.Chat.ID] = game

		nmsg, err := gifMessageReply(
			msg.Chat.ID, "startgame", config.DefaultImages.Start,
			msg, bot, user, joinButton(msg.Chat.ID, bot),
		)
		if err != nil {
			gameList[msg.Chat.ID] = nil
			return err
		}
		game.MsgSent.StartMsg = nmsg.MessageID

		plMsg, err := markdownMessage(msg.Chat.ID, "players", bot, game.Users)
		if err != nil {
			gameList[msg.Chat.ID] = nil
			return err
		}
		game.MsgSent.PlayerList = plMsg.MessageID
		return startGamePM(msg, bot)
	}
	if _, err := markdownReply(int64(msg.From.ID), "grouponly", msg, bot, nil); err != nil {
		return err
	}
	return nil
}

func onHelp(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot == nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID < 0 {
		if _, err := markdownReply(int64(msg.Chat.ID), "help", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
	if _, err := markdownReply(int64(msg.From.ID), "help", msg, bot, nil); err != nil {
		return err
	}
	return nil
}

func onReceiveAnimation(msg *tgApi.Message, other ...interface{}) error {
	if !isAdminMode {
		log.Println("No admin mode on, skip")
		return nil
	}
	filename := strings.ToLower(strings.SplitN(msg.Document.FileName, ".", 2)[0])
	log.Printf("Got filename %s", filename)
	fileID := msg.Document.FileID
	switch filename {
	case "win":
		config.DefaultImages.Win = fileID
	case "lose":
		config.DefaultImages.Lose = fileID
	case "startgame":
		fallthrough
	case "start":
		config.DefaultImages.Start = fileID
	case "killed":
		config.DefaultImages.Killed = fileID
	case "trapped":
		config.DefaultImages.Trapped = fileID
	}
	appPath, _ := filepath.Abs(path.Dir(os.Args[0]))
	log.Printf("Find config in `%s`\n", appPath)
	if err := config.DefaultImages.WriteConfig(path.Join(appPath, "images.dat")); err != nil {
		return err
	}
	return nil
}

func onAdmin(msg *tgApi.Message, other ...interface{}) error {
	bot, args := messageUtils(other)
	if bot != nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID < 0 {
		if _, err := markdownReply(msg.Chat.ID, "chatonly", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
	if args == "" {
		isAdminMode = false
		if _, err := markdownReply(int64(msg.From.ID), "adminoff", msg, bot, nil); err != nil {
			return err
		}
		return nil
	} else if args == config.DefaultConfig.AdminPassword {
		isAdminMode = true
		if _, err := markdownReply(int64(msg.From.ID), "adminon", msg, bot, nil); err != nil {
			return err
		}
		return nil
	} else {
		if _, err := markdownReply(int64(msg.From.ID), "padpassword", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
}

func onExtend(msg *tgApi.Message, other ...interface{}) error {
	bot, args := messageUtils(other)
	if bot != nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID > 0 {
		if _, err := markdownReply(msg.Chat.ID, "grouponly", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
	eta, err := strconv.Atoi(args)
	if err != nil {
		eta = 30
	}
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	langSet := getLang(msg.Chat.ID)
	game := gameList[msg.Chat.ID]
	game.TimeLeft += eta
	if game.TimeLeft > 300 {
		game.TimeLeft = 300
	}

	nmsg := tgApi.NewMessage(msg.Chat.ID, lang.T(langSet, "jointime",
		fmt.Sprintf("%d:%d", game.TimeLeft/60, game.TimeLeft%60),
	))
	nmsg.ReplyMarkup = joinButton(msg.Chat.ID, bot)
	m, err := bot.Send(nmsg)
	if err != nil {
		return err
	}
	game.MsgSent.JoinTimeMsg = append(game.MsgSent.JoinTimeMsg, m.MessageID)
	return nil
}

func onPlayers(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot != nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID > 0 {
		if _, err := markdownReply(msg.Chat.ID, "grouponly", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	game := gameList[msg.Chat.ID]
	if game != nil {
		if _, err := markdownReply(msg.Chat.ID, "onplayers", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func onFlee(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot != nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID > 0 {
		if _, err := markdownReply(msg.Chat.ID, "grouponly", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
	user, err := getUser(int64(msg.From.ID))
	if err != nil {
		return err
	}
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	game := gameList[msg.Chat.ID]
	if game != nil {
		if game.Status == models.GameNotStart {
			game.Flee(user)
			if _, err := markdownMessage(msg.Chat.ID, "flee", bot, user); err != nil {
				return err
			}
			return nil
		}
		if _, err := markdownMessage(msg.Chat.ID, "noflee", bot, nil); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func onSetLang(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot != nil {
		return errors.New("no bot found")
	}

	btns := make([]tgApi.InlineKeyboardButton, 0)
	for l := range basis.GlobalLanguageList {
		btns = append(btns, tgApi.NewInlineKeyboardButtonData(l, "setlang="+l))
	}
	var mk tgApi.InlineKeyboardMarkup
	for i := 0; i < len(btns); i += 2 {
		if len(btns)-i <= len(btns)%2 {
			mk.InlineKeyboard = append(mk.InlineKeyboard, btns[i:])
		} else {
			mk.InlineKeyboard = append(mk.InlineKeyboard, btns[i:i+2])
		}
	}
	if msg.Chat.ID > 0 {
		if _, err := markdownReply(int64(msg.From.ID), "setlang", msg, bot, nil, mk); err != nil {
			return err
		}
	} else if msg.Chat.ID < 0 {
		if _, err := markdownReply(msg.Chat.ID, "setlang", msg, bot, nil, mk); err != nil {
			return err
		}
	}
	return nil
}

func onStat(msg *tgApi.Message, other ...interface{}) error {

	bot, args := messageUtils(other)
	if bot != nil {
		return errors.New("no bot found")
	}
	var (
		userID int64
		err    error
	)
	if args == "" {
		userID = int64(msg.From.ID)
	} else {
		userID, err = strconv.ParseInt(args, 10, 64)
		if err != nil {
			userID = int64(msg.From.ID)
		}
	}
	user, err := getUser(userID)
	if err != nil {
		return err
	}
	if _, err := markdownMessage(msg.Chat.ID, "userstats", bot, user); err != nil {
		return err
	}
	return nil
}

func onNextGame(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot != nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID > 0 {
		if _, err := markdownReply(msg.Chat.ID, "grouponly", msg, bot, nil); err != nil {
			return err
		}
		return nil
	}
	user, err := getUser(int64(msg.From.ID))
	if err != nil {
		return err
	}
	var gameQueue []int64
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueFormatString, msg.Chat.ID),
	).Scan(&gameQueue); err != nil {
		return err
	}
	gameQueue = append(gameQueue, user.TgUserID)
	if err := database.Redis.Set(
		fmt.Sprintf(consts.GameQueueFormatString, msg.Chat.ID),
		gameQueue, -1,
	).Err(); err != nil {
		return err
	}
	langSet := getLang(int64(msg.From.ID))
	reply := tgApi.NewMessage(
		int64(msg.From.ID),
		lang.T(langSet, "gamequeue", msg.Chat.Title),
	)
	btn := tgApi.NewInlineKeyboardButtonData(
		lang.T(langSet, "cancel", nil),
		fmt.Sprintf("cancelgame=%d", msg.Chat.ID),
	)
	reply.ReplyMarkup = tgApi.NewInlineKeyboardMarkup(tgApi.NewInlineKeyboardRow(btn))
	nmsg, err := bot.Send(reply)
	if err != nil {
		return err
	}
	var gameQueueMsg []int
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueMsgFormatString, msg.Chat.ID),
	).Scan(&gameQueueMsg); err != nil {
		return err
	}
	gameQueueMsg = append(gameQueueMsg, nmsg.MessageID)
	if err := database.Redis.Set(
		fmt.Sprintf(consts.GameQueueMsgFormatString, msg.Chat.ID),
		gameQueueMsg, -1,
	).Err(); err != nil {
		return err
	}
	return nil
}

func onForceStart(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot == nil {
		return errors.New("no bot found")
	}
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	if msg.Chat.ID < 0 {
		game := gameList[msg.Chat.ID]
		if game != nil {
			if err := game.Start(); err != nil {
				if _, err := markdownMessage(msg.Chat.ID, "noenoughplayers", bot, nil); err != nil {
					return err
				}
				return nil
			}
		}
	}
	if _, err := markdownMessage(msg.Chat.ID, "grouponly", bot, nil); err != nil {
		return err
	}
	return nil
}
