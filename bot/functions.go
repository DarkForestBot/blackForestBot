package bot

import (
	"fmt"

	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func MarkdownReply(to int64, key string, replyTo int, bot *Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(to)
	reply := tgApi.NewMessage(to, lang.T(langSet, key, other))
	reply.ReplyToMessageID = msg.MessageID
	reply.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func MarkdownMessage(to int64, key string, bot *Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(to)
	reply := tgApi.NewMessage(to, lang.T(langSet, key, other))
	reply.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func GifReply(to int64, key string, image string, replyTo int, bot *Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(to)
	reply := tgApi.NewDocumentShare(to, image)
	reply.Caption = lang.T(langSet, key, other)
	reply.MimeType = "video/mp4"
	reply.ParseMode = tgApi.ModeMarkdown
	reply.ReplyToMessageID = replyTo
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func GifMessage(to int64, key string, image string, bot *Bot, other interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	langSet := getLang(to)
	reply := tgApi.NewDocumentShare(to, image)
	reply.Caption = lang.T(langSet, key, other)
	reply.MimeType = "video/mp4"
	reply.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		reply.ReplyMarkup = mk[0]
	}
	return bot.Send(reply)
}

func JoinButton(tgGroupID int64, bot *Bot) tgApi.InlineKeyboardMarkup {
	langSet := getLang(tgGroupID)
	joinButton := tgApi.NewInlineKeyboardButtonURL(
		lang.T(langSet, "join", nil),
		fmt.Sprintf("https://t.me/%s?start=%d", bot.Name(), tgGroupID),
	)
	return tgApi.NewInlineKeyboardMarkup(tgApi.NewInlineKeyboardRow(joinButton))
}

func StartGamePM(msg *tgApi.Message, bot *Bot) error {
	var gameQueue []int64
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueFormatString, msg.Chat.ID),
	).Scan(&gameQueue); err != nil {
		return err
	}
	for _, i := range gameQueue {
		MarkdownMessage(i, "newgame", bot, msg.Chat.Title)
	}
	return nil
}
