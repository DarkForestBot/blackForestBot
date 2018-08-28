package controllers

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.wetofu.top/tonychee7000/blackForestBot/bot"
	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var gameList map[int64]*models.Game

func init() {
	gameList = make(map[int64]*models.Game)
}

func onJoinAChat(msg *tgApi.Message, other ...interface{}) error {
	if msg.Chat.ID > 0 {
		return errors.New("Bad tg group")
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
		if err := database.DB.Save(group).Error; err != nil {
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
	if args != "" {
		// join game
	}

	return nil
}

func onStartGame(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot == nil {
		return errors.New("no bot found")
	}
	if msg.Chat.ID < 0 {
		langSet := getLang(msg.Chat.ID)
		user, err := getUser(int64(msg.From.ID))
		if err != nil {
			return err
		}
		group, err := getTgGroup(msg.Chat.ID)
		if err != nil {
			return err
		}
		game := new(models.Game)
		game.TgGroup = group
		gameList[msg.Chat.ID] = game

		reply := tgApi.NewDocumentShare(msg.Chat.ID, config.DefaultImages.Start)
		reply.ReplyToMessageID = msg.MessageID
		reply.Caption = lang.T(langSet, "startgame", user)
		reply.MimeType = "video/mp4"
		joinButton := tgApi.NewInlineKeyboardButtonURL(
			lang.T(langSet, "join", nil),
			fmt.Sprintf("https://t.me/%s?start=%d", bot.Name(), msg.Chat.ID),
		)
		reply.ReplyMarkup = tgApi.NewInlineKeyboardMarkup(tgApi.NewInlineKeyboardRow(joinButton))
		if _, err := bot.Send(reply); err != nil {
			return err
		}
	} else {
		langSet := getLang(int64(msg.From.ID))
		reply := tgApi.NewMessage(msg.Chat.ID, lang.T(langSet, "grouponly", nil))
		reply.ReplyToMessageID = msg.MessageID
		reply.ParseMode = tgApi.ModeMarkdown
		if _, err := bot.Send(reply); err != nil {
			return err
		}
	}
	return nil
}

func onHelp(msg *tgApi.Message, other ...interface{}) error {
	bot, _ := messageUtils(other)
	if bot == nil {
		return errors.New("no bot found")
	}
	var langSet string
	if msg.Chat.ID < 0 {
		langSet = getLang(int64(msg.Chat.ID))
	} else {
		langSet = getLang(int64(msg.From.ID))
	}
	reply := tgApi.NewMessage(msg.Chat.ID, lang.T(langSet, "help", nil))
	reply.ReplyToMessageID = msg.MessageID
	reply.ParseMode = tgApi.ModeMarkdown
	_, err := bot.Send(reply)
	return err
}

func onReceiveAnimation(msg *tgApi.Message, other ...interface{}) error {
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

func messageUtils(args []interface{}) (*bot.Bot, string) {
	var (
		Bot     *bot.Bot
		argText string
		ok      bool
	)
	for _, o := range args {
		if Bot == nil {
			if Bot, ok = o.(*bot.Bot); ok {
				continue
			}
		}
		if argText == "" {
			if argText, ok = o.(string); ok {
				continue
			}
		}
	}
	return Bot, argText
}

func getLang(ID int64) string {
	var str string
	if err := database.Redis.Get(fmt.Sprintf(consts.LangSetFormatString, ID)).Scan(&str); err != nil {
		log.Printf("WARNING: error %v", err)
		return "English"
	}
	return str
}

func getUser(ID int64) (*models.User, error) {
	user := new(models.User)
	if err := database.DB.Where(models.User{TgUserID: ID}).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func getTgGroup(ID int64) (*models.TgGroup, error) {
	group := new(models.TgGroup)
	if err := database.DB.Where(models.TgGroup{TgGroupID: ID}).First(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}
