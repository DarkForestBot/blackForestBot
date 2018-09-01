package models

//List of channels
var (
	NewGameHint           chan *Game
	UserJoinHint          chan *User
	GameFleeHint          chan *User
	GameNoFleeHint        chan *User
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
	PlayersHint           chan *Game // Change player list
	PlayerKillHint        chan *Player
	PlayerBeastHint       chan *Player
	GetPlayersHint        chan *Game
	UserStatsHint         chan *User
	ShootXHint            chan *Player
	ShootYHint            chan *Player
	UnionReqHint          chan []*Player //Player[0]: Src, Player[1]: Dst
	UnionAcceptHint       chan []*Player
	UnionRejectHint       chan []*Player
)

func init() {
	NewGameHint = make(chan *Game, 1024)
	UserJoinHint = make(chan *User, 1024)
	GameFleeHint = make(chan *User, 1024)
	GameNoFleeHint = make(chan *User, 1024)
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
	GetPlayersHint = make(chan *Game, 1024)
	UserStatsHint = make(chan *User, 1024)
	ShootXHint = make(chan *Player, 1024)
	ShootYHint = make(chan *Player, 1024)
	UnionReqHint = make(chan []*Player, 1024)
	UnionAcceptHint = make(chan []*Player, 1024)
	UnionRejectHint = make(chan []*Player, 1024)
}
