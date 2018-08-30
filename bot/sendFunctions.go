package bot

import (
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) makeMessage(to int64, langSet, key string,
	obj interface{}, mk ...tgApi.InlineKeyboardMarkup) tgApi.MessageConfig {
	msg := tgApi.NewMessage(to, lang.T(langSet, key, obj))
	msg.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		msg.ReplyMarkup = mk[0]
	}
	return msg
}

//MarkdownReply is
func (b *Bot) MarkdownReply(to int64, langSet, key string, replyTo int,
	obj interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	reply := b.makeMessage(to, langSet, key, obj, mk...)
	reply.ReplyToMessageID = replyTo
	return b.Send(reply)
}

//MarkdownMessage is
func (b *Bot) MarkdownMessage(to int64, langSet, key string,
	obj interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	return b.Send(b.makeMessage(to, langSet, key, obj, mk...))
}

func (b *Bot) makeGifMessage(to int64, langSet, key, imageID string,
	obj interface{}, mk ...tgApi.InlineKeyboardMarkup) tgApi.DocumentConfig {
	doc := tgApi.NewDocumentShare(to, imageID)
	doc.Caption = lang.T(langSet, key, obj)
	doc.MimeType = "video/mp4"
	doc.ParseMode = tgApi.ModeMarkdown
	if mk != nil && len(mk) == 1 {
		doc.ReplyMarkup = mk[0]
	}
	return doc
}

//GifReply is
func (b *Bot) GifReply(to int64, langSet, key, imageID string, replyTo int,
	obj interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	reply := b.makeGifMessage(to, langSet, key, imageID, obj, mk...)
	reply.ReplyToMessageID = replyTo
	return b.Send(reply)
}

//GifMessage is
func (b *Bot) GifMessage(to int64, langSet, key, imageID string,
	obj interface{}, mk ...tgApi.InlineKeyboardMarkup) (tgApi.Message, error) {
	return b.Send(b.makeGifMessage(to, langSet, key, imageID, obj, mk...))
}
