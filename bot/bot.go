package bot

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ProceedFunc recv and reply
type ProceedFunc func(tgApi.Update, *Bot) error

// Bot a bot
type Bot struct {
	updateConfig   tgApi.UpdateConfig
	updatesChannel tgApi.UpdatesChannel
	bot            *tgApi.BotAPI
	procFunc       ProceedFunc
}

// NewBot for create
func NewBot() *Bot {
	var b = new(Bot)
	b.updatesChannel = make(tgApi.UpdatesChannel, 1024)
	b.updateConfig = tgApi.NewUpdate(0)
	return b
}

// Connect to bot
func (b *Bot) Connect(conf config.Config) error {
	var err error
	b.bot, err = tgApi.NewBotAPI(conf.APIToken)
	if err != nil {
		return err
	}
	b.bot.Debug = conf.Debug
	b.updateConfig.Timeout = conf.UpdateTimeout

	b.updatesChannel, err = b.bot.GetUpdatesChan(b.updateConfig)

	if err != nil {
		return err
	}
	return nil
}

// Run to proceed
func (b *Bot) Run() {
	for {
		select {
		case update := <-b.updatesChannel:
			err := b.procFunc(update, b)
			if err != nil {
				log.Println("ERROR:", err)
			}
		}
	}
}

// Send message to tg
func (b *Bot) Send(c tgApi.Chattable) (tgApi.Message, error) {
	return b.bot.Send(c)
}

// AnswerCallbackQuery wrapper
func (b *Bot) AnswerCallbackQuery(c tgApi.CallbackConfig) (tgApi.APIResponse, error) {
	return b.bot.AnswerCallbackQuery(c)
}

// RegisterProcessor to reg some func.
func (b *Bot) RegisterProcessor(fn ProceedFunc) {
	b.procFunc = fn
}

// Name of bot
func (b *Bot) Name() string {
	return b.bot.Self.UserName
}

// ID of bot
func (b *Bot) ID() int {
	return b.bot.Self.ID
}
