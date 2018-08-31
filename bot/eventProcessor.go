package bot

import (
	"fmt"
	"log"
	"sync"

	"git.wetofu.top/tonychee7000/blackForestBot/basis"
	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"git.wetofu.top/tonychee7000/blackForestBot/controllers"
	"git.wetofu.top/tonychee7000/blackForestBot/database"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) onNewGameHint(game *models.Game) {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	langSet := game.TgGroup.Lang
	var (
		msg tgApi.Message
		err error
	)
	//Step I: send game start message.
	msg, err = b.GifMessage(
		game.TgGroup.TgGroupID, langSet, "startgame",
		config.DefaultImages.Start, game.Founder,
		joinButton(game.TgGroup.TgGroupID, b),
	)
	if err != nil {
		log.Println("ERROR:", err)
	}
	game.MsgSent.StartMsg = msg.MessageID

	//Step II: show players joined.
	msg, err = b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "players", game,
	)
	if err != nil {
		log.Println("ERROR:", err)
	}
	game.MsgSent.PlayerList = msg.MessageID

	//Setp III: pm in gamequeue
	var gameQueue []int64
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueFormatString, game.TgGroup.TgGroupID),
	).Scan(&gameQueue); err != nil {
		log.Println("ERROR:", err)
		return
	}
	for _, id := range gameQueue {
		langSet := getLang(id)
		_, err = b.MarkdownMessage(id, langSet, "newgame", game.TgGroup.Name)
		if err != nil {
			log.Println("ERROR:", err)
		}
	}
}

func (b *Bot) onUserJoinHint(user *models.User) {
	langSet := user.TgGroupJoinGame.Lang
	//Step I: Join game hint in group
	if _, err := b.MarkdownMessage(
		user.TgGroupJoinGame.TgGroupID, langSet, "joingame", user,
	); err != nil {
		log.Println("ERROR:", err)
	}
	//Setp II: PM user join success.
	langSet = user.Language
	if _, err := b.MarkdownMessage(
		user.TgUserID, langSet, "joingame", user.TgGroupJoinGame,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGameFleeHint(user *models.User) {
	langSet := user.TgGroupJoinGame.Lang
	if _, err := b.MarkdownMessage(
		user.TgGroupJoinGame.TgGroupID, langSet, "flee", user,
	); err != nil {
		log.Println("ERROR", err)
	}
}

func (b *Bot) onGameNoFleeHint(user *models.User) {
	langSet := user.TgGroupJoinGame.Lang
	if _, err := b.MarkdownMessage(
		user.TgGroupJoinGame.TgGroupID, langSet, "noflee", user,
	); err != nil {
		log.Println("ERROR", err)
	}
}

func (b *Bot) onNotEnoughPlayersHint(game *models.Game) {
	langSet := game.TgGroup.Lang
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "noenoughplayers", nil,
	); err != nil {
		log.Println("ERROR", err)
	}
}

func (b *Bot) onJoinTimeLeftHint(game *models.Game) {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	langSet := game.TgGroup.Lang
	msg, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "jointime",
		fmt.Sprintf("%d:%d", game.TimeLeft/60, game.TimeLeft%60),
		joinButton(game.TgGroup.TgGroupID, b),
	)
	if err != nil {
		log.Println("ERROR:", err)
	}
	game.MsgSent.JoinTimeMsg = append(game.MsgSent.JoinTimeMsg, msg.MessageID)
}

func (b *Bot) onStartGameFailed(game *models.Game) {
	b.startGameClearMessage(game)
	langSet := game.TgGroup.Lang
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "gamecancelled", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onStartGameSuccess(game *models.Game) {
	b.startGameClearMessage(game)
	langSet := game.TgGroup.Lang
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "gamestart", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAbortPlayerHint(player *models.Player) {
	langSet := player.User.Language
	if _, err := b.MarkdownMessage(
		player.User.TgUserID, langSet, "timeout", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGameChangeToDayHint(game *models.Game) {
	langSet := game.TgGroup.Lang
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "onday", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
	// send everyone operations...
}

func (b *Bot) onGameChangeToNightHint(game *models.Game) {
	langSet := game.TgGroup.Lang
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "onnight", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
	// send everyone operations...
}

func (b *Bot) onPlayersHint(game *models.Game) {
	controllers.EditMessageTextEvent <- tgApi.NewEditMessageText(
		game.TgGroup.TgGroupID, game.MsgSent.PlayerList,
		lang.T(game.TgGroup.Lang, "players", game),
	)
}

func (b *Bot) onPlayerKillHint(player *models.Player) {
	var (
		image   string
		langSet string
	)
	if player.KilledReason == models.Trapped {
		image = config.DefaultImages.Trapped
	} else {
		image = config.DefaultImages.Killed
	}
	langSet = player.User.Language
	if _, err := b.GifMessage(
		player.User.TgUserID, langSet,
		fmt.Sprintf("killed%d", player.KilledReason),
		image, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onPlayerBeastHint(player *models.Player) {
	langSet := player.User.Language
	if _, err := b.MarkdownMessage(
		player.User.TgUserID, langSet, "beast", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGameLoseHint(game *models.Game) {
	langSet := game.TgGroup.Lang
	if _, err := b.GifMessage(
		game.TgGroup.TgGroupID, langSet, config.DefaultImages.Lose,
		"lose", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onWinGameHint(game *models.Game) {
	langSet := game.TgGroup.Lang
	if _, err := b.GifMessage(
		game.TgGroup.TgGroupID, langSet, config.DefaultImages.Win,
		"win", game.Winner,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGetPlayersHint(game *models.Game) {
	langSet := game.TgGroup.Lang
	if _, err := b.MarkdownReply(
		game.TgGroup.TgGroupID, langSet, "onplayers",
		game.MsgSent.PlayerList, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onUserStatsHint(user *models.User) {
	langSet := getLang(user.QueryMsg.Chat.ID)
	if _, err := b.MarkdownReply(
		user.QueryMsg.Chat.ID, langSet, "userstats",
		user.QueryMsg.MessageID, user,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onJoinAChatEvent(msg *tgApi.Message) {
	langSet := "English"
	if _, err := b.MarkdownMessage(
		msg.Chat.ID, langSet, "joinachat", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onReceiveAnimationEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "receivegif",
		msg.MessageID, msg.Document.FileName,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onPMOnlyEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "chatonly",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGroupOnlyEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "grouponly",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onStartEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownMessage(
		msg.Chat.ID, langSet, "onstart", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGroupHasAGameEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "hasgame",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onHelpEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownMessage(
		msg.Chat.ID, langSet, "help", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAboutEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "about",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAdminModeOffEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "adminoff",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAdminModeOnEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "adminon",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAdminBadPasswordEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "badpassword",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onSetLangMsgEvent(msg *tgApi.Message) {
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
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownReply(msg.Chat.ID, langSet, "setlang", msg.MessageID, nil, mk); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onNextGameEvent(msg *tgApi.Message) {
	langSet := getLang(int64(msg.From.ID))
	nmsg, err := b.MarkdownMessage(
		int64(msg.From.ID), langSet, "gamequeue", msg.Chat.Title,
		tgApi.NewInlineKeyboardMarkup(
			tgApi.NewInlineKeyboardRow(
				tgApi.NewInlineKeyboardButtonData(
					lang.T(langSet, "cancel", nil),
					fmt.Sprintf("cancelgame=%d", msg.Chat.ID),
				),
			),
		),
	)
	if err != nil {
		log.Println("ERROR:", err)
	}
	var gameQueueMsg []int
	if err := database.Redis.Get(
		fmt.Sprintf(consts.GameQueueMsgFormatString, msg.Chat.ID),
	).Scan(&gameQueueMsg); err != nil {
		log.Println("ERROR:", err)
	}
	gameQueueMsg = append(gameQueueMsg, nmsg.MessageID)
	if err := database.Redis.Set(
		fmt.Sprintf(consts.GameQueueMsgFormatString, msg.Chat.ID),
		gameQueueMsg, -1,
	).Err(); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onLanguageChangedEvent(act *tgApi.CallbackQuery) {
	langSet := getLang(int64(act.Message.Chat.ID))
	if _, err := b.MarkdownMessage(
		act.Message.Chat.ID, langSet, "langchanged", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onDeleteMessageEvent(c tgApi.DeleteMessageConfig) {
	if _, err := b.DeleteMessage(c); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onRemoveMessageMarkUpEvent(c tgApi.EditMessageReplyMarkupConfig) {
	defer func() {
		x := recover()
		if x != nil {
			log.Println("ERROR:", x)
		}
	}()
	c.ReplyMarkup = nil
	if _, err := b.Send(c); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onEditMessageTextEvent(c tgApi.EditMessageTextConfig) {
	if _, err := b.Send(c); err != nil {
		log.Println("ERROR:", err)
	}
}
