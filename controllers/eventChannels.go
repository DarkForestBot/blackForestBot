package controllers

import tgApi "github.com/go-telegram-bot-api/telegram-bot-api"

//List of message event
var (
	OnJoinAChatEvent         chan *tgApi.Message
	OnReceiveAnimationEvent  chan *tgApi.Message
	PMOnlyEvent              chan *tgApi.Message
	OnStartEvent             chan *tgApi.Message
	GroupHasAGameEvent       chan *tgApi.Message
	GroupOnlyEvent           chan *tgApi.Message
	HelpEvent                chan *tgApi.Message
	AboutEvent               chan *tgApi.Message
	AdminModeOffEvent        chan *tgApi.Message
	AdminModeOnEvent         chan *tgApi.Message
	AdminBadPasswordEvent    chan *tgApi.Message
	SetLangMsgEvent          chan *tgApi.Message
	NextGameEvent            chan *tgApi.Message
	RegisterNeededEvent      chan *tgApi.Message
	LanguageChangedEvent     chan *tgApi.CallbackQuery
	DeleteMessageEvent       chan tgApi.DeleteMessageConfig
	RemoveMessageMarkUpEvent chan tgApi.EditMessageReplyMarkupConfig
	EditMessageTextEvent     chan tgApi.EditMessageTextConfig
)

func init() {
	OnJoinAChatEvent = make(chan *tgApi.Message, 1024)
	OnReceiveAnimationEvent = make(chan *tgApi.Message, 1024)
	PMOnlyEvent = make(chan *tgApi.Message, 1024)
	OnStartEvent = make(chan *tgApi.Message, 1024)
	GroupHasAGameEvent = make(chan *tgApi.Message, 1024)
	GroupOnlyEvent = make(chan *tgApi.Message, 1024)
	HelpEvent = make(chan *tgApi.Message, 1024)
	AboutEvent = make(chan *tgApi.Message, 1024)
	AdminModeOffEvent = make(chan *tgApi.Message, 1024)
	AdminModeOnEvent = make(chan *tgApi.Message, 1024)
	AdminBadPasswordEvent = make(chan *tgApi.Message, 1024)
	SetLangMsgEvent = make(chan *tgApi.Message, 1024)
	NextGameEvent = make(chan *tgApi.Message, 1024)
	RegisterNeededEvent = make(chan *tgApi.Message, 1024)
	LanguageChangedEvent = make(chan *tgApi.CallbackQuery, 1024)
	DeleteMessageEvent = make(chan tgApi.DeleteMessageConfig, 1024)
	RemoveMessageMarkUpEvent = make(chan tgApi.EditMessageReplyMarkupConfig, 1024)
	EditMessageTextEvent = make(chan tgApi.EditMessageTextConfig, 1024)
}
