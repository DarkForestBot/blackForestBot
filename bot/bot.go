package bot

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ProceedFunc recv and reply
type ProceedFunc func(tgApi.Update, *Bot)

// Bot a bot
type Bot struct {
	updateConfig   tgApi.UpdateConfig
	updatesChannel tgApi.UpdatesChannel
	bot            *tgApi.BotAPI
	procFunc       ProceedFunc
}

//DefaultBot and only one bot here
var DefaultBot *Bot

func init() {
	DefaultBot = NewBot()
	if err := DefaultBot.Connect(config.DefaultConfig); err != nil {
		log.Fatalln("FATAL:", err)
	}
	log.Printf("Bot authoirzed by name: %s(%d)", DefaultBot.Name(), DefaultBot.ID())
	DefaultBot.RegisterProcessor(messageProcessor)
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
	go b.messageManager()
	for {
		select {
		case update := <-b.updatesChannel:
			go b.procFunc(update, b)
		}
	}
}

// Send message to tg
func (b *Bot) Send(c tgApi.Chattable) (tgApi.Message, error) {
	return b.bot.Send(c)
}

// DeleteMessage wrapper
func (b *Bot) DeleteMessage(config tgApi.DeleteMessageConfig) (tgApi.APIResponse, error) {
	return b.bot.DeleteMessage(config)
}

// AnswerCallbackQuery wrapper
func (b *Bot) AnswerCallbackQuery(c tgApi.CallbackConfig) (tgApi.APIResponse, error) {
	return b.bot.AnswerCallbackQuery(c)
}

// LeaveChat wrapper
func (b *Bot) LeaveChat(c tgApi.ChatConfig) (tgApi.APIResponse, error) {
	return b.bot.LeaveChat(c)
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
