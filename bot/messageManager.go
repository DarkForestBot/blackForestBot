package bot

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/controllers"
	"git.wetofu.top/tonychee7000/blackForestBot/models"
)

//just used for send out messages
func (b *Bot) messageManager() {
	log.Println("messageManager is running.")
	for {
		select {
		case user := <-models.AchivementRewardedHint:
			go b.onAchivementRewardedHint(user)
		case game := <-models.NewGameHint: // startmsg, playerlist, nextgameuerpm
			go b.onNewGameHint(game)
		case user := <-models.UserJoinHint:
			go b.onUserJoinHint(user)
		case user := <-models.GameFleeHint:
			go b.onGameFleeHint(user)
		case user := <-models.GameNoFleeHint:
			go b.onGameNoFleeHint(user)
		case game := <-models.NotEnoughPlayersHint:
			go b.onNotEnoughPlayersHint(game)
		case game := <-models.JoinTimeLeftHint:
			go b.onJoinTimeLeftHint(game)
		case game := <-models.StartGameFailed:
			go b.onStartGameFailed(game)
		case game := <-models.StartGameSuccess: // clear next game pm list
			go b.onStartGameSuccess(game)
		case game := <-models.GameTimeOutOperation: // edit some messages or other staff
			go b.onGameTimeOutOperation(game)
		case player := <-models.AbortPlayerHint:
			go b.onAbortPlayerHint(player)
		case game := <-models.GameChangeToDayHint:
			go b.onGameChangeToDayHint(game)
		case game := <-models.GameChangeToNightHint:
			go b.onGameChangeToNightHint(game)
		case player := <-models.ShootXHint:
			go b.onShootXHint(player)
		case player := <-models.ShootYHint:
			go b.onShootYHint(player)
		case players := <-models.UnionReqHint:
			go b.onUnionReqHint(players)
		case players := <-models.UnionAcceptHint:
			go b.onUnionAcceptHint(players)
		case players := <-models.UnionRejectHint:
			go b.onUnionRejectHint(players)
		case game := <-models.GameLoseHint:
			go b.onGameLoseHint(game)
		case game := <-models.WinGameHint:
			go b.onWinGameHint(game)
		case game := <-models.PlayersHint: // Change player list before start
			go b.onPlayersHint(game)
		case player := <-models.PlayerKillHint:
			go b.onPlayerKillHint(player)
		case player := <-models.PlayerBeastHint:
			go b.onPlayerBeastHint(player)
		case game := <-models.GetPlayersHint:
			go b.onGetPlayersHint(game)
		case user := <-models.UserStatsHint:
			go b.onUserStatsHint(user)
		case msg := <-controllers.OnJoinAChatEvent:
			go b.onJoinAChatEvent(msg)
		case msg := <-controllers.OnReceiveAnimationEvent:
			go b.onReceiveAnimationEvent(msg)
		case msg := <-controllers.PMOnlyEvent:
			go b.onPMOnlyEvent(msg)
		case msg := <-controllers.OnStartEvent:
			go b.onStartEvent(msg)
		case msg := <-controllers.GroupHasAGameEvent:
			go b.onGroupHasAGameEvent(msg)
		case msg := <-controllers.GroupOnlyEvent:
			go b.onGroupOnlyEvent(msg)
		case msg := <-controllers.HelpEvent:
			go b.onHelpEvent(msg)
		case msg := <-controllers.AboutEvent:
			go b.onAboutEvent(msg)
		case msg := <-controllers.AdminModeOffEvent:
			go b.onAdminModeOffEvent(msg)
		case msg := <-controllers.AdminModeOnEvent:
			go b.onAdminModeOnEvent(msg)
		case msg := <-controllers.AdminBadPasswordEvent:
			go b.onAdminBadPasswordEvent(msg)
		case msg := <-controllers.SetLangMsgEvent:
			go b.onSetLangMsgEvent(msg)
		case msg := <-controllers.NextGameEvent:
			go b.onNextGameEvent(msg)
		case act := <-controllers.LanguageChangedEvent:
			go b.onLanguageChangedEvent(act)
		case deleteConf := <-controllers.DeleteMessageEvent:
			go b.onDeleteMessageEvent(deleteConf)
		case edit := <-controllers.RemoveMessageMarkUpEvent:
			go b.onRemoveMessageMarkUpEvent(edit)
		case edit := <-controllers.EditMessageTextEvent:
			go b.onEditMessageTextEvent(edit)
		}
	}
}
