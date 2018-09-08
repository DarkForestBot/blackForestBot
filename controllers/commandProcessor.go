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

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type eventFunc func(*tgApi.Message, ...string) error

var isAdminMode bool

//CommandList is
var CommandList map[string]eventFunc

func init() {
	CommandList = make(map[string]eventFunc)
	CommandList["start"] = onStart
	CommandList["help"] = onHelp
	CommandList["about"] = onAbout
	CommandList["startgame"] = onStartGame
	CommandList["admin"] = onAdmin
	CommandList["extend"] = onExtend
	CommandList["players"] = onPlayers
	CommandList["flee"] = onFlee
	CommandList["setlang"] = onSetLang
	CommandList["stats"] = onStat
	CommandList["forcestart"] = onForceStart
	CommandList["nextgame"] = onNextGame
	CommandList["newgame"] = onStartGame
	isAdminMode = false
}

//OnJoinAChat is
func OnJoinAChat(msg *tgApi.Message) error {
	if msg.Chat.ID > 0 {
		return errors.New("Bad tg group")
	}

	group := new(models.TgGroup)
	if err := database.DB.Where(models.TgGroup{TgGroupID: msg.Chat.ID}).Assign(
		models.TgGroup{
			Name: msg.Chat.Title,
			Admin: &models.User{
				TgUserID:   int64(msg.From.ID),
				Name:       fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName),
				TgUserName: msg.From.UserName,
			},
			Active: true,
		},
	).FirstOrCreate(group).Error; err != nil {
		return err
	}
	lang.UserLang[group.TgGroupID] = group.Lang

	log.Printf("Group `%s` registered.\n", group.Name)
	OnJoinAChatEvent <- msg
	return nil
}

//OnKickoutAChat is
func OnKickoutAChat(msg *tgApi.Message) error {
	if msg.Chat.ID > 0 {
		return errors.New("Bad tg group")
	}
	group, err := models.GetTgGroup(msg.Chat.ID)
	if err != nil {
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

//OnReceiveAnimation is
func OnReceiveAnimation(msg *tgApi.Message) error {
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
	case "beast":
		config.DefaultImages.Beast = fileID
	}
	appPath, _ := filepath.Abs(path.Dir(os.Args[0]))
	log.Printf("Find config in `%s`\n", appPath)
	if err := config.DefaultImages.WriteConfig(path.Join(appPath, "images.dat")); err != nil {
		return err
	}
	OnReceiveAnimationEvent <- msg
	return nil
}

func onStart(msg *tgApi.Message, args ...string) error {
	if msg.Chat.ID < 0 {
		PMOnlyEvent <- msg
		return errors.New("pm only")
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
	lang.UserLang[user.TgUserID] = user.Language
	log.Printf("User `%s(%d)` registered.\n", user.Name, user.TgUserID)

	if args != nil && len(args) != 0 && args[0] != "" {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}
		game, ok := gameList[id]
		if ok && game != nil && game.Status == models.GameNotStart {
			if game.GetUser(user.TgUserID) == nil {
				game.Join(user)
			}
		}
	} else {
		OnStartEvent <- msg
	}

	return nil
}

func onStartGame(msg *tgApi.Message, args ...string) error {
	if msg.Chat.ID < 0 {
		user, err := models.GetUser(int64(msg.From.ID))
		if err != nil {
			return err
		}
		group, err := models.GetTgGroup(msg.Chat.ID)
		if err != nil {
			return err
		}

		game, ok := gameList[msg.Chat.ID]
		if ok && game != nil && game.TgGroup.TgGroupID == msg.Chat.ID &&
			game.Status != models.GameOver {
			GroupHasAGameEvent <- msg
		} else {
			game = models.NewGame(group, user)
			gameList[msg.Chat.ID] = game
		}
	} else {
		GroupOnlyEvent <- msg
	}
	return nil
}

func onHelp(msg *tgApi.Message, args ...string) error {
	HelpEvent <- msg
	return nil
}

func onAbout(msg *tgApi.Message, args ...string) error {
	AboutEvent <- msg
	return nil
}

func onAdmin(msg *tgApi.Message, args ...string) error {
	if msg.Chat.ID < 0 {
		PMOnlyEvent <- msg
		return nil
	}
	if args == nil || len(args) == 0 || args[0] == "" {
		isAdminMode = false
		AdminModeOffEvent <- msg
		return nil
	}
	if args != nil && len(args) == 1 && args[0] == config.DefaultConfig.AdminPassword {
		isAdminMode = true
		AdminModeOnEvent <- msg
		return nil
	}
	AdminBadPasswordEvent <- msg
	return nil
}

func onExtend(msg *tgApi.Message, args ...string) error {
	if msg.Chat.ID > 0 {
		GroupOnlyEvent <- msg
		return errors.New("Group only")
	}
	var (
		eta int
		err error
	)
	if args == nil || len(args) == 0 {
		eta = 30
	} else {
		eta, err = strconv.Atoi(args[0])
		if err != nil {
			eta = 30
		}
	}

	game, ok := gameList[msg.Chat.ID]
	if ok && game != nil {
		game.Extend(eta)
	} else {
		NoGameEvent <- msg
	}
	return nil
}

func onPlayers(msg *tgApi.Message, args ...string) error {
	if msg.Chat.ID > 0 {
		GroupOnlyEvent <- msg
		return errors.New("Group only")
	}

	game, ok := gameList[msg.Chat.ID]
	if ok && game != nil {
		game.HintPlayers()
	} else {
		NoGameEvent <- msg
	}
	return nil
}

func onFlee(msg *tgApi.Message, args ...string) error {
	if msg.Chat.ID > 0 {
		GroupOnlyEvent <- msg
		return errors.New("Group only")
	}

	game, ok := gameList[msg.Chat.ID]
	if ok && game != nil {
		user := game.GetUser(int64(msg.From.ID))
		if user == nil {
			return errors.New("No such user found")
		}
		game.Flee(user)
	} else {
		NoGameEvent <- msg
	}
	return nil
}

func onSetLang(msg *tgApi.Message, args ...string) error {
	SetLangMsgEvent <- msg
	return nil
}

func onStat(msg *tgApi.Message, args ...string) error {
	var (
		userName string
		err      error
	)
	if args == nil || len(args) == 0 || args[0] == "" {
		userName = msg.From.UserName
	} else {
		userName = strings.Replace(args[0], "@", "", 1)
	}
	user, err := models.GetUserByUserName(userName)
	if err != nil {
		return err
	}
	if user != nil {
		user.Stats(msg)
	}
	return nil
}

func onNextGame(msg *tgApi.Message, arg ...string) error {
	if msg.Chat.ID > 0 {
		GroupOnlyEvent <- msg
		return errors.New("Group only")
	}
	_, err := models.GetUser(int64(msg.From.ID))
	if err != nil {
		return err
	}

	NextGameEvent <- msg
	return nil
}

func onForceStart(msg *tgApi.Message, arg ...string) error {
	if msg.Chat.ID < 0 {
		game, ok := gameList[msg.Chat.ID]
		if ok && game != nil {
			return game.ForceStart()
		}
		NoGameEvent <- msg
		return errors.New("No game found")
	}
	GroupOnlyEvent <- msg
	return errors.New("Group only")
}
