package models

//List of channels
var (
	NewGameHint           chan *Game
	UserJoinHint          chan *User
	GameFleeHint          chan *User
	NotEnoughPlayersHint  chan *Game
	JoinTimeLeftHint      chan *Game
	StartGameFailed       chan *Game
	StartGameSuccess      chan *Game
	GameTimeOutOperation  chan *Game
	AbortPlayerHint       chan *Player
	GameChangeToNightHint chan *Game
	GameChangeToDayHint   chan *Game
	GameLoseHint          chan *Game
	WinGameHint           chan *Game
	PlayersHint           chan *Game
	PlayerKillHint        chan *Player
	PlayerBeastHint       chan *Player
)

func init() {
	NewGameHint = make(chan *Game, 1024)
	UserJoinHint = make(chan *User, 1024)
	GameFleeHint = make(chan *User, 1024)
	NotEnoughPlayersHint = make(chan *Game, 1024)
	JoinTimeLeftHint = make(chan *Game, 1024)
	StartGameFailed = make(chan *Game, 1024)
	StartGameSuccess = make(chan *Game, 1024)
	GameTimeOutOperation = make(chan *Game, 1024)
	AbortPlayerHint = make(chan *Player, 1024)
	GameChangeToNightHint = make(chan *Game, 1024)
	GameChangeToDayHint = make(chan *Game, 1024)
	GameLoseHint = make(chan *Game, 1024)
	WinGameHint = make(chan *Game, 1024)
	PlayersHint = make(chan *Game, 1024)
	PlayerKillHint = make(chan *Player, 1024)
	PlayerBeastHint = make(chan *Player, 1024)
}
