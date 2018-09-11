package bot

import (
	"fmt"
	"log"
	"sync"

	"git.wetofu.top/tonychee7000/blackForestBot/basis"
	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/controllers"
	"git.wetofu.top/tonychee7000/blackForestBot/lang"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) onAchivementRewardedHint(userAchivement models.UserAchivement) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	defer func() { recover() }()
	if config.DefaultConfig.Debug {
		log.Println("onAchivementRewardedHint user=", userAchivement["userid"], "code=", userAchivement["achivementcode"])
	}
	langSet := getLang(userAchivement["userid"])
	_, err := b.MarkdownMessage(
		userAchivement["userid"], langSet, "achivementrewarded",
		lang.T(
			langSet, fmt.Sprintf(
				"achivement%03d", userAchivement["achivementcode"]), nil),
	)
	if err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onNewGameHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	langSet := getLang(game.TgGroup.TgGroupID)
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
	gameQueue, err := controllers.GetGameQueue(game.TgGroup.TgGroupID)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}
	for _, e := range gameQueue {
		langSet := getLang(e.UserID)
		_, err = b.MarkdownMessage(e.UserID, langSet, "newgame", game.TgGroup.Name)
		if err != nil {
			log.Println("ERROR:", err)
		}
	}
}

func (b *Bot) onUserJoinHint(user *models.User) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(user.TgGroupJoinGame.TgGroupID)
	//Step I: Join game hint in group
	if _, err := b.MarkdownMessage(
		user.TgGroupJoinGame.TgGroupID, langSet, "joingame", user,
	); err != nil {
		log.Println("ERROR:", err)
	}
	//Setp II: PM user join success.
	langSet = getLang(user.TgUserID)
	if _, err := b.MarkdownMessage(
		user.TgUserID, langSet, "joinsuccess", user.TgGroupJoinGame,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGameFleeHint(user *models.User) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(user.TgGroupJoinGame.TgGroupID)
	if _, err := b.MarkdownMessage(
		user.TgGroupJoinGame.TgGroupID, langSet, "flee", user,
	); err != nil {
		log.Println("ERROR", err)
	}
}

func (b *Bot) onGameNoFleeHint(user *models.User) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(user.TgGroupJoinGame.TgGroupID)
	if _, err := b.MarkdownMessage(
		user.TgGroupJoinGame.TgGroupID, langSet, "noflee", user,
	); err != nil {
		log.Println("ERROR", err)
	}
}

func (b *Bot) onNotEnoughPlayersHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "noenoughplayers", nil,
	); err != nil {
		log.Println("ERROR", err)
	}
}

func (b *Bot) onJoinTimeLeftHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	langSet := getLang(game.TgGroup.TgGroupID)
	msg, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "jointime",
		fmt.Sprintf("%02d:%02d", game.TimeLeft/60, game.TimeLeft%60),
		joinButton(game.TgGroup.TgGroupID, b),
	)
	if err != nil {
		log.Println("ERROR:", err)
	}
	game.MsgSent.JoinTimeMsg = append(game.MsgSent.JoinTimeMsg, msg.MessageID)
}

func (b *Bot) onTryStartGameHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		game.TgGroup.TgGroupID, game.MsgSent.StartMsg, tgApi.InlineKeyboardMarkup{},
	)
	for _, id := range game.MsgSent.JoinTimeMsg {
		controllers.DeleteMessageEvent <- tgApi.DeleteMessageConfig{
			ChatID:    game.TgGroup.TgGroupID,
			MessageID: id,
		}
	}
}

func (b *Bot) onStartGameFailed(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	b.startGameClearMessage(game)

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "gamecancelled", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onStartGameSuccess(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	b.startGameClearMessage(game)

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "gamestart", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
	for _, player := range game.Players {
		langSet := getLang(player.User.TgUserID)
		if _, err := b.MarkdownMessage(
			player.User.TgUserID, langSet, "getposition", player,
		); err != nil {
			log.Println("ERROR:", err)
		}
	}
	// Clear nextgame queue
	if err := controllers.ClearGameQueue(game.TgGroup.TgGroupID); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGameTimeOutOperation(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	for _, player := range game.Players {
		if !player.Live {
			continue
		}
		if player.OperationMsg != 0 {
			controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
				player.User.TgUserID, player.OperationMsg,
				tgApi.InlineKeyboardMarkup{},
			)
			player.OperationMsg = 0
		}
		if player.UnionReq != 0 {
			controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
				player.User.TgUserID, player.UnionReq,
				tgApi.InlineKeyboardMarkup{},
			)
			player.UnionReq = 0
		}
		if len(player.UnionReqRecv) != 0 {
			for _, msg := range player.UnionReqRecv {
				langSet := getLang(msg.From.User.TgUserID)
				controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
					player.User.TgUserID, msg.Msg.MessageID,
					tgApi.InlineKeyboardMarkup{},
				)
				if _, err := b.MarkdownMessage(
					msg.From.User.TgUserID, langSet, "unionfailed", nil,
				); err != nil {
					log.Println("ERROR:", err)
				}
			}
			player.UnionReqRecv = []*models.UnionReqRecv{}
		}
		if game.IsDay {
			langSet := getLang(player.User.TgUserID)
			if _, err := b.MarkdownMessage(
				player.User.TgUserID, langSet, "timeoutday", nil,
			); err != nil {
				log.Println("ERROR:", err)
			}
		}
	}
}

func (b *Bot) onAbortPlayerHint(player *models.Player) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		player.User.TgUserID, player.OperationMsg,
		tgApi.InlineKeyboardMarkup{},
	)
	langSet := getLang(player.User.TgUserID)
	if _, err := b.MarkdownMessage(
		player.User.TgUserID, langSet, "timeoutnight", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGameChangeToDayHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "onday", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "gameplayers", game,
	); err != nil {
		log.Println("ERROR:", err)
	}
	// send everyone operations...
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	// check any one can union?
	var availList = make([]*models.Player, 0)
	for _, player := range game.Players {
		if player.Live && !player.UnionValidation() {
			availList = append(availList, player)
		}
	}
	if len(availList)-1 == 0 { //except self.
		return
	}
	// send msg
	for _, player := range game.Players {
		if player.Live && !player.UnionValidation() {
			langSet = getLang(player.User.TgUserID)
			msg, err := b.MarkdownMessage(
				player.User.TgUserID, langSet, "unionhint", nil,
				makeUnionButtons(game, player),
			)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
			player.UnionReq = msg.MessageID
		}
	}
}

func (b *Bot) onGameChangeToNightHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.MarkdownMessage(
		game.TgGroup.TgGroupID, langSet, "onnight", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
	// send everyone operations...
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	for _, player := range game.Players {
		if player.Live {
			langSet := getLang(player.User.TgUserID)
			if player.ShootNoneStreak+1 >= models.PlayerShootNoneStreakLimit {
				if _, err := b.MarkdownMessage(
					player.User.TgUserID, langSet, "strikeout", game,
				); err != nil {
					log.Println("ERROR:", err)
				}
				game.AttachOperation(player.Abort())
				continue
			}
			if _, err := b.MarkdownMessage(
				player.User.TgUserID, langSet, "gameplayers", game,
			); err != nil {
				log.Println("ERROR:", err)
			}
			msg, err := b.MarkdownMessage(
				player.User.TgUserID, langSet, "operhint1", nil,
				makeNightOperations(game.TgGroup.TgGroupID, player, len(game.Players), 0),
			)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
			player.OperationMsg = msg.MessageID
		}
	}
}

func (b *Bot) onShootXHint(player *models.Player) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		player.User.TgUserID, player.OperationMsg,
		tgApi.InlineKeyboardMarkup{},
	)
	if player.Live {
		langSet := getLang(player.User.TgUserID)
		msg, err := b.MarkdownMessage(
			player.User.TgUserID, langSet, "operhint2", nil,
			makeNightOperations(
				player.User.TgGroupJoinGame.TgGroupID, player,
				player.CurrentGamePlayersCount, 1),
		)
		if err != nil {
			log.Println("ERROR:", err)
		}
		player.OperationMsg = msg.MessageID
	}
}

func (b *Bot) onShootYHint(player *models.Player) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		player.User.TgUserID, player.OperationMsg,
		tgApi.InlineKeyboardMarkup{},
	)
	player.OperationMsg = 0 // Clear message Operation
}

func (b *Bot) onUnionReqHint(players []*models.Player) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(players[1].User.TgUserID)
	mk := tgApi.NewInlineKeyboardMarkup(
		tgApi.NewInlineKeyboardRow(
			tgApi.NewInlineKeyboardButtonData(
				lang.T(langSet, "accept", nil),
				fmt.Sprintf(
					"unionaccept=%d,%d", players[0].User.TgUserID,
					players[0].User.TgGroupJoinGame.TgGroupID,
				),
			),
			tgApi.NewInlineKeyboardButtonData(
				lang.T(langSet, "reject", nil),
				fmt.Sprintf(
					"unionreject=%d,%d", players[0].User.TgUserID,
					players[0].User.TgGroupJoinGame.TgGroupID,
				),
			),
		),
	)

	// Step I: remove union request button.
	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		players[0].User.TgUserID, players[0].UnionReq,
		tgApi.InlineKeyboardMarkup{},
	)

	// Step II: req has sent
	if _, err := b.MarkdownMessage(
		players[0].User.TgUserID, getLang(players[0].User.TgUserID),
		"unionreqsent", players[1].User,
	); err != nil {
		log.Println("ERROR:", err)
	}

	// Step III: send request
	nmsg, err := b.MarkdownMessage(
		players[1].User.TgUserID, langSet, "unionreq", players[0].User, mk,
	)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}

	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	players[1].UnionReqRecv = append(players[1].UnionReqRecv, models.NewUnionReqRecv(nmsg, players[0]))
}

func (b *Bot) onUnionAcceptHint(players []*models.Player) {
	defer func() { recover() }()
	threadLimitPool <- 1
	defer releaseThreadPool()

	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		players[0].User.TgUserID, players[0].UnionReq, tgApi.InlineKeyboardMarkup{},
	)
	controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
		players[1].User.TgUserID, players[1].UnionReq, tgApi.InlineKeyboardMarkup{},
	)

	langSet := getLang(players[0].User.TgUserID)
	// 0: button clicker, 1: reply to
	if _, err := b.MarkdownMessage(
		players[0].User.TgUserID, langSet, "unionsuccess", players[1].User,
	); err != nil {
		log.Println("ERROR:", err)
	}

	langSet = getLang(players[1].User.TgUserID)
	if _, err := b.MarkdownMessage(
		players[1].User.TgUserID, langSet, "unionsuccess", players[0].User,
	); err != nil {
		log.Println("ERROR:", err)
	}

	for _, msg := range players[0].UnionReqRecv {
		controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
			msg.Msg.Chat.ID, msg.Msg.MessageID, tgApi.InlineKeyboardMarkup{},
		)
		if msg.From != players[1] {
			langSet := getLang(msg.From.User.TgUserID)
			if _, err := b.MarkdownMessage(
				msg.From.User.TgUserID, langSet, "unionfailed", nil,
			); err != nil {
				log.Println("ERROR:", err)
			}
		}
	}
	for _, msg := range players[1].UnionReqRecv {
		controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
			msg.Msg.Chat.ID, msg.Msg.MessageID, tgApi.InlineKeyboardMarkup{},
		)
		if msg.From != players[0] {
			langSet := getLang(msg.From.User.TgUserID)
			if _, err := b.MarkdownMessage(
				msg.From.User.TgUserID, langSet, "unionfailed", nil,
			); err != nil {
				log.Println("ERROR:", err)
			}
		}
	}
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	players[0].UnionReqRecv = []*models.UnionReqRecv{}
	players[1].UnionReqRecv = []*models.UnionReqRecv{}
}

func (b *Bot) onUnionRejectHint(players []*models.Player) {
	defer func() { recover() }()
	threadLimitPool <- 1
	defer releaseThreadPool()
	// 0: button clicker, 1: reply to
	langSet := getLang(players[1].User.TgUserID)
	if _, err := b.MarkdownMessage(
		players[1].User.TgUserID, langSet, "unionfailed", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}

	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	for i, msg := range players[0].UnionReqRecv {
		if msg.From == players[1] {
			controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
				msg.Msg.Chat.ID, msg.Msg.MessageID, tgApi.InlineKeyboardMarkup{},
			)
			//Remove this msg from UnionReqRecv
			if i == len(players[0].UnionReqRecv)-1 { // if the last one
				players[0].UnionReqRecv = players[0].UnionReqRecv[:i] // remove last one
			} else {
				players[0].UnionReqRecv = append(players[0].UnionReqRecv[:i], players[0].UnionReqRecv[i+1:]...)
			}
		}
	}
}

func (b *Bot) onUnionHasOneHint(players []*models.Player) {
	defer func() { recover() }()
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(players[0].User.TgUserID)
	if _, err := b.MarkdownMessage(
		players[0].User.TgUserID, langSet, "unionhasone", players[1].User,
	); err != nil {
		log.Println("ERROR:", err)
	}

	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	for i, msg := range players[0].UnionReqRecv {
		if msg.From == players[1] {
			controllers.RemoveMessageMarkUpEvent <- tgApi.NewEditMessageReplyMarkup(
				msg.Msg.Chat.ID, msg.Msg.MessageID, tgApi.InlineKeyboardMarkup{},
			)
			//Remove this msg from UnionReqRecv
			if i == len(players[0].UnionReqRecv)-1 { // if the last one
				players[0].UnionReqRecv = players[0].UnionReqRecv[:i] // remove last one
			} else {
				players[0].UnionReqRecv = append(players[0].UnionReqRecv[:i], players[0].UnionReqRecv[i+1:]...)
			}
		}
	}
}

func (b *Bot) onPlayersHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(game.TgGroup.TgGroupID)
	controllers.EditMessageTextEvent <- tgApi.NewEditMessageText(
		game.TgGroup.TgGroupID, game.MsgSent.PlayerList,
		lang.T(langSet, "players", game),
	)
}

func (b *Bot) onPlayerKillHint(player *models.Player) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	var image string

	if player.KilledReason == models.Trapped {
		image = config.DefaultImages.Trapped
	} else {
		image = config.DefaultImages.Killed
	}

	langSet := getLang(player.User.TgUserID)
	if _, err := b.GifMessage(
		player.User.TgUserID, langSet,
		fmt.Sprintf("killed%d", player.KilledReason),
		image, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onPlayerBeastHint(player *models.Player) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(player.User.TgUserID)
	if _, err := b.GifMessage(
		player.User.TgUserID, langSet, "beast",
		config.DefaultImages.Beast, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGameLoseHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.GifMessage(
		game.TgGroup.TgGroupID, langSet, "lose",
		config.DefaultImages.Lose, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
	if err := b.makeReplay(game); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onWinGameHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.GifMessage(
		game.TgGroup.TgGroupID, langSet, "win",
		config.DefaultImages.Win, game.Winner.User,
	); err != nil {
		log.Println("ERROR:", err)
	}
	if err := b.makeReplay(game); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGetPlayersHint(game *models.Game) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(game.TgGroup.TgGroupID)
	if _, err := b.MarkdownReply(
		game.TgGroup.TgGroupID, langSet, "onplayers",
		game.MsgSent.PlayerList, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onUserStatsHint(user *models.User) {
	threadLimitPool <- 1
	defer releaseThreadPool()

	langSet := getLang(user.QueryMsg.Chat.ID)
	if _, err := b.MarkdownReply(
		user.QueryMsg.Chat.ID, langSet, "userstats",
		user.QueryMsg.MessageID, user,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onJoinAChatEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownMessage(
		msg.Chat.ID, langSet, "joinachat", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onReceiveAnimationEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "receivegif",
		msg.MessageID, msg.Document.FileName,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onPMOnlyEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "chatonly",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGroupOnlyEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "grouponly",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onStartEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownMessage(
		msg.Chat.ID, langSet, "onstart", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onGroupHasAGameEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "hasgame",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onHelpEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownMessage(
		msg.Chat.ID, langSet, "help", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAboutEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "about",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAdminModeOffEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "adminoff",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAdminModeOnEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "adminon",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onAdminBadPasswordEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(int64(msg.Chat.ID))
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "badpassword",
		msg.MessageID, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onSetLangMsgEvent(msg *tgApi.Message) {
	threadLimitPool <- 1
	defer releaseThreadPool()
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
	threadLimitPool <- 1
	defer releaseThreadPool()
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
	if err := controllers.AddGameQueue(
		msg.Chat.ID, models.NewQueueElement(
			int64(msg.From.ID), nmsg.MessageID)); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onLanguageChangedEvent(act *tgApi.CallbackQuery) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	langSet := getLang(int64(act.Message.Chat.ID))
	if _, err := b.MarkdownMessage(
		act.Message.Chat.ID, langSet, "langchanged", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onDeleteMessageEvent(c tgApi.DeleteMessageConfig) {
	threadLimitPool <- 1
	defer releaseThreadPool()
	if _, err := b.DeleteMessage(c); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onRemoveMessageMarkUpEvent(c tgApi.EditMessageReplyMarkupConfig) {
	threadLimitPool <- 1
	defer releaseThreadPool()
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
	threadLimitPool <- 1
	defer releaseThreadPool()
	c.ParseMode = tgApi.ModeMarkdown
	if _, err := b.Send(c); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onRegisterNeededEvent(msg *tgApi.Message) {
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "registerneeded",
		msg.MessageID, nil); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onNoGameEvent(msg *tgApi.Message) {
	langSet := getLang(msg.Chat.ID)
	if _, err := b.MarkdownReply(
		msg.Chat.ID, langSet, "nogame",
		msg.MessageID, nil); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onOperationApprovedEvent(act *tgApi.CallbackQuery) {
	langSet := getLang(int64(act.From.ID))
	if _, err := b.MarkdownMessage(
		int64(act.From.ID), langSet, "operapproved", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onShootApprovedHint(operation *models.Operation) {
	langSet := getLang(operation.Player.User.TgUserID)
	if operation.Target == nil {
		log.Println("ERROR: no target found.")
		return
	}
	var msgHint string
	if operation.Player.Status < models.PlayerStatusBeast {
		msgHint = "shootapproved"
	} else {
		msgHint = "shootapproved_beast"
	}
	if _, err := b.MarkdownMessage(
		operation.Player.User.TgUserID, langSet,
		msgHint, operation.Target,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onPlayerSurvivedAtNightHint(player *models.Player) {
	langSet := getLang(player.User.TgUserID)
	if _, err := b.MarkdownMessage(
		player.User.TgUserID, langSet, "survive", nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onPlayerShootNothingHint(player *models.Player) {
	langSet := getLang(player.User.TgUserID)
	var msgHint string
	if player.Status >= models.PlayerStatusBeast {
		msgHint = "eatnothing"
	} else {
		msgHint = "shootnothing"
	}
	if _, err := b.MarkdownMessage(
		player.User.TgUserID, langSet, msgHint, nil,
	); err != nil {
		log.Println("ERROR:", err)
	}
}

func (b *Bot) onPlayerShootSomethingHint(player *models.Player) {
	langSet := getLang(player.User.TgUserID)
	if player.Target == nil {
		log.Println("ERROR: no target.")
		return
	}
	var msgHint string
	if player.Status >= models.PlayerStatusBeast {
		msgHint = "eatsomething"
	} else {
		msgHint = "shootsomething"
	}
	if _, err := b.MarkdownMessage(
		player.User.TgUserID, langSet, msgHint, player,
	); err != nil {
		log.Println("ERROR:", err)
	}
}
