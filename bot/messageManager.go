package bot

import (
	"log"

	"git.wetofu.top/tonychee7000/blackForestBot/models"
)

//just used for send out messages
func (b *Bot) messageManager() {
	log.Println("messageManager is running.")
	for {
		select {
		case game := <-models.NewGameHint: // startmsg, playerlist, nextgameuerpm
		case user := <-models.UserJoinHint:
		case user := <-models.GameFleeHint:
		case user := <-models.GameNoFleeHint:
		case game := <-models.NotEnoughPlayersHint:
		case game := <-models.JoinTimeLeftHint:
		case game := <-models.StartGameFailed:
		case game := <-models.StartGameSuccess: // clear next game pm list
		case game := <-models.GameTimeOutOperation:
		case player := <-models.AbortPlayerHint:
		case game := <-models.GameChangeToDayHint:
		case game := <-models.GameChangeToNightHint:
		case game := <-models.GameLoseHint:
		case game := <-models.WinGameHint:
		case game := <-models.PlayersHint:
		case player := <-models.PlayerKillHint:
		case player := <-models.PlayerBeastHint:
		case game := <-models.GetPlerysHint:
		case user := <-models.UserStatsHint:
		}
	}
}
