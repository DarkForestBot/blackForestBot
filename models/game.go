package models

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"git.wetofu.top/tonychee7000/blackForestBot/config"
	"git.wetofu.top/tonychee7000/blackForestBot/consts"
	"github.com/robfig/cron"
)

//GameStatus is
type GameStatus int

//List of game status
const (
	GameNotStart   GameStatus = 0
	GameStart      GameStatus = 1
	GameOver       GameStatus = 2
	GameIsDay                 = true
	GameIsNight               = !GameIsDay
	GameMinPlayers            = 6
)

type msgSent struct {
	StartMsg     int
	PlayerList   int
	JoinTimeMsg  []int
	UnionOperMsg []int
}

func newMsgSent() *msgSent {
	m := new(msgSent)
	m.JoinTimeMsg = make([]int, 0)
	m.UnionOperMsg = make([]int, 0)
	return m
}

type globalOperation struct {
	Player    *Player
	Action    PlayerAction
	Target    *Position // When result might nil
	IsResult  bool
	EachOther bool

	//Result
	Killed   *Player
	BeKilled bool
	BeBeast  bool
	Survive  bool // for Abort
	None     bool

	Finally bool
}

//Game is
type Game struct {
	Round            int
	IsDay            bool
	Founder          *User
	Users            []*User
	Status           GameStatus
	Positions        []*Position
	Players          []*Player
	TgGroup          *TgGroup
	TimeLeft         int
	MsgSent          *msgSent
	Operations       []*Operation // Every round will clear
	GlobalOperations [][]globalOperation
	Cron             *cron.Cron
	Winner           *Player
}

//NewGame is to create a new game in the group
func NewGame(tg *TgGroup, founder *User) *Game {
	game := new(Game)
	game.Founder = founder
	game.Round = 0
	game.IsDay = GameIsDay
	game.Users = make([]*User, 0)
	game.Status = GameNotStart
	game.Positions = make([]*Position, 0)
	game.Players = make([]*Player, 0)
	game.TgGroup = tg
	game.TimeLeft = consts.TwoMinutes
	game.MsgSent = newMsgSent()

	game.Cron = cron.New()
	game.Cron.AddFunc("@every 1s", game.RunCheck)
	game.Cron.Start()

	NewGameHint <- game
	return game
}

//Extend is to extend join time
func (g *Game) Extend(timeSecond int) {
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	if timeSecond < 0 {
		timeSecond = -timeSecond
	}

	if g.Status == GameNotStart {
		g.TimeLeft += timeSecond
		if g.TimeLeft >= consts.FiveMinutes {
			g.TimeLeft = consts.FiveMinutes
		}
	} else {
		return
	}
	JoinTimeLeftHint <- g
}

//Join is add user to game
func (g *Game) Join(user *User) {
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	if g.Status == GameNotStart {
		//Don't join game repeat.
		for _, gu := range g.Users {
			if gu == user {
				return
			}
		}
		g.Users = append(g.Users, user)
		user.TgGroupJoinGame = g.TgGroup

		if g.TimeLeft < consts.OneMinute {
			g.TimeLeft += 10
			if g.TimeLeft > consts.OneMinute {
				g.TimeLeft = consts.OneMinute
			}
		}
		UserJoinHint <- user
		PlayersHint <- g
	}
}

//Flee is remove user to game or kill player in game
func (g *Game) Flee(user *User) {
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()

	switch g.Status {
	case GameNotStart:
		g.fleeUser(user)
		GameFleeHint <- user
		PlayersHint <- g
	case GameStart:
		GameNoFleeHint <- user
	}
}

//Start is go!
func (g *Game) Start() error {
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	debug := 1
	if g.Status != GameNotStart {
		return nil
	}

	if config.DefaultConfig.Debug {
		debug++
	}
	if len(g.Users) < GameMinPlayers/debug {
		return errors.New("Too less users")
	}
	g.makePlayer()
	g.Status = GameStart
	g.TimeLeft = consts.TwoMinutes
	g.IsDay = GameIsDay
	g.Round = 1 // First day
	GameChangeToDayHint <- g
	for _, p := range g.Players {
		if config.DefaultConfig.Debug {
			log.Println("DEBUG:", p)
		}
	}
	return nil
}

//ForceStart is
func (g *Game) ForceStart() error {
	if err := g.Start(); err != nil {
		NotEnoughPlayersHint <- g
		return err
	}
	StartGameSuccess <- g
	return nil
}

func (g *Game) String() string {
	return fmt.Sprintf(
		"Game(round=%d, isday=%v, timeleft=%d, tgId=%d)",
		g.Round, g.IsDay, g.TimeLeft, g.TgGroup.TgGroupID,
	)
}

//GetPlayer is
func (g *Game) GetPlayer(tgUserID int64) *Player {
	for _, player := range g.Players {
		if player.User.TgUserID == tgUserID {
			return player
		}
	}
	return nil
}

//GetUser is
func (g *Game) GetUser(tgUserID int64) *User {
	for _, user := range g.Users {
		if user.TgUserID == tgUserID {
			return user
		}
	}
	return nil
}

//GetPosition is
func (g *Game) GetPosition(x, y int) *Position {
	for _, pos := range g.Positions {
		if pos.CheckPosition(x, y) {
			return pos
		}
	}
	return nil
}

// AttachOperation is
func (g *Game) AttachOperation(op *Operation) {
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	// One player only operate once a round.
	for _, operation := range g.Operations {
		if op.Player == operation.Player {
			return
		}
	}
	g.Operations = append(g.Operations, op)
}

//HintPlayers is called by command /players
func (g *Game) HintPlayers() {
	GetPlayersHint <- g
}

// RunCheck is
func (g *Game) RunCheck() {
	g.countDown()
	switch g.Status {
	case GameNotStart:
		g.joinTimeCheck()
	case GameStart:
		g.gameTimeCheck()
	case GameOver:
		fallthrough
	default:
		g.Cron.Stop()
	}
}

func (g *Game) joinTimeCheck() {
	switch g.TimeLeft {
	case 60:
		fallthrough
	case 30:
		fallthrough
	case 10:
		JoinTimeLeftHint <- g
	case 0:
		TryStartGameHint <- g
		if g.Status != GameNotStart {
			return
		}
		if err := g.Start(); err != nil {
			var lock sync.RWMutex
			lock.Lock()
			defer lock.Unlock()
			g.Status = GameOver
			StartGameFailed <- g
		} else {
			StartGameSuccess <- g
		}
	}
}

func (g *Game) gameTimeCheck() {
	if g.TimeLeft != 0 {
		var c = 0
		for _, p := range g.Players {
			if p.Live {
				c++
			}
		}
		if g.IsDay || len(g.Operations) != c {
			return
		}
	}
	//Step I: editMessage sent.
	GameTimeOutOperation <- g

	//Step II: check who has no action and sent abort()
	for _, p := range g.findAbort() {
		if !g.IsDay {
			g.AttachOperation(p.Abort())
			AbortPlayerHint <- p
		}
	}
	//Step III: check and change phase
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	if g.IsDay {
		g.IsDay = GameIsNight
		g.TimeLeft = consts.OneMinute
		GameChangeToNightHint <- g
	} else {
		g.settle()
		g.IsDay = GameIsDay
		g.TimeLeft = consts.TwoMinutes
		g.Round++
		if !g.winloseCheck() {
			GameChangeToDayHint <- g
		}
	}
}

func (g *Game) winloseCheck() bool {
	var (
		pl     = make([]*Player, 0)
		status = false
	)
	for _, player := range g.Players {
		if player.Live {
			pl = append(pl, player)
		}
	}
	log.Println("Current live:", pl)
	if len(pl) == 0 { // All Dead
		g.GlobalOperations[len(g.GlobalOperations)-1] = append(
			g.GlobalOperations[len(g.GlobalOperations)-1],
			globalOperation{
				Finally:  true,
				IsResult: true,
			},
		)
		GameLoseHint <- g
		status = true
	} else if len(pl) == 1 {
		pl[0].User.GamesWon++
		pl[0].User.Update()
		g.Winner = pl[0]
		WinGameHint <- g
		g.GlobalOperations[len(g.GlobalOperations)-1] = append(
			g.GlobalOperations[len(g.GlobalOperations)-1],
			globalOperation{
				Player:   g.Winner,
				Finally:  true,
				IsResult: true,
			},
		)
		status = true
	}
	if status {
		g.Status = GameOver
		for _, player := range g.Players {
			player.User.CheckAchivement()
		}
	}
	return status
}

// core logic!
func (g *Game) settle() {
	// Stage 0: Correction Union.
	for _, player := range g.Players {
		player.UnionCorrection()
	}
	var operations = make([]globalOperation, 0)
	// Stage I: tag the target and beast the player abort
	g.settleStageTag(&operations)
	// Stage II: check betray.
	g.settleStageCheckBetry(&operations)
	// Stage III: check who surely dead.
	g.settleStageCheckDeath(&operations)
	// Stage IV: check trap
	g.settleStageCheckTrap(&operations)
	// Stage V: check union
	g.settleStageCheckUnion(&operations)
	// Stage VI: expose position
	g.settleStageExposePosition(&operations)
	// Stage O: Reset some status
	for _, player := range g.Players {
		player.ActionClear()
		player.User.Update()
	}
	g.Operations = make([]*Operation, 0)
	g.GlobalOperations = append(g.GlobalOperations, operations)
}

func (g *Game) settleStageTag(gop *[]globalOperation) {
	for _, operation := range g.Operations {
		(*gop) = append(*gop, globalOperation{
			Player: operation.Player,
			Action: operation.Action,
			Target: operation.Target,
		})
		switch operation.Action {
		case Shoot: // Betray is special shoot
			if operation.Target != nil && operation.Target.Player != nil {
				operation.Player.Target = operation.Target.Player
			} else {
				operation.Player.User.ShootCount++
				PlayerShootNothingHint <- operation.Player
				(*gop) = append(*gop, globalOperation{
					Player:   operation.Player,
					Action:   operation.Action,
					IsResult: true,
				})
			}
		case Abort:
			rand.Seed(time.Now().Unix())
			if operation.Player.Status < PlayerStatusBeast {
				fate := rand.Intn(4) // 1/n rate
				if fate == 0 {
					if operation.Player.Status < PlayerStatusBeast {
						(*gop) = append(*gop, globalOperation{
							Player:   operation.Player,
							Action:   operation.Action,
							BeBeast:  true,
							IsResult: true,
						})
					}
					operation.Player.StatusChange(PlayerStatusBeast)
				} else if fate == 1 {
					if operation.Player.Live {
						(*gop) = append(*gop, globalOperation{
							Player:   operation.Player,
							Action:   operation.Action,
							BeKilled: true,
							IsResult: true,
						})
					}
					operation.Player.Kill(Flee)
				} else {
					PlayerSurvivedAtNightHint <- operation.Player
					(*gop) = append(*gop, globalOperation{
						Player:   operation.Player,
						Action:   operation.Action,
						Survive:  true,
						IsResult: true,
					})
				}
			} else {
				if rand.Intn(3) == 0 { // 1/n
					if operation.Player.Live {
						(*gop) = append(*gop, globalOperation{
							Player:   operation.Player,
							Action:   operation.Action,
							BeKilled: true,
							IsResult: true,
						})
					}
					operation.Player.Kill(Flee)
				} else {
					PlayerSurvivedAtNightHint <- operation.Player
					(*gop) = append(*gop, globalOperation{
						Player:   operation.Player,
						Action:   operation.Action,
						Survive:  true,
						IsResult: true,
					})
				}
			}
		}
	}
}

func (g *Game) settleStageCheckBetry(gop *[]globalOperation) {
	for _, player := range g.Players {
		if !player.UnionValidation() { // Here must some mistakes.
			player.Ununion()
			continue
		}
		if player.Target != nil && player.Target == player.Unioned { // Betray
			player.User.BetrayCount++
			if player.Unioned.Target == player { // Betray each other
				player.Unioned.User.BetrayCount++
				if player.Status < PlayerStatusBeast {
					(*gop) = append(*gop, globalOperation{
						Player:    player,
						Action:    Betray,
						IsResult:  true,
						EachOther: true,
						BeBeast:   true,
					})
				}
				player.StatusChange(PlayerStatusBeast)
				if player.Unioned.Status < PlayerStatusBeast {
					(*gop) = append(*gop, globalOperation{
						Player:    player.Unioned,
						Action:    Betray,
						IsResult:  true,
						EachOther: true,
						BeBeast:   true,
					})
				}
				player.Unioned.StatusChange(PlayerStatusBeast)
				player.Target = nil
				player.Unioned.Target = nil
				player.Ununion() // Union broken.
			} else { // I betrayed my union
				if player.Target.TrapSet {
					if player.Live {
						(*gop) = append(*gop, globalOperation{
							Player:   player,
							Action:   Betray,
							BeKilled: true,
							IsResult: true,
						})
					}
					player.Kill(Trapped) // Oops! I was trapped!
					player.User.KilledByTrapCount++
				} else {
					if player.Target.Live {
						(*gop) = append(*gop, globalOperation{
							Player:   player,
							Action:   Betray,
							Killed:   player.Target,
							IsResult: true,
						})
					}
					player.Target.Kill(Betrayed)
					player.Target = nil
				}
			}
		}
	}
}

func (g *Game) settleStageCheckDeath(gop *[]globalOperation) {
	for _, player := range g.Players {
		if player.Target != nil && player.Target.Live { // I want to kill some one.
			if player.Status >= PlayerStatusBeast { // I am a beast
				if player.Target.Target == player { // Kill each other
					if player.Target.Status >= PlayerStatusBeast { // That is a beast also...NO!!
						PlayerShootSomethingHint <- player.Target //Will hint target that player dead
						PlayerShootSomethingHint <- player        //Will hint player that target dead
						if player.Target.Live && player.Live {
							(*gop) = append(*gop, globalOperation{
								Player:    player,
								Action:    Shoot,
								EachOther: true,
								Killed:    player.Target,
								BeKilled:  true,
								IsResult:  true,
							})
						}
						player.Target.Kill(BeastKill)
						player.Kill(BeastKill) // All dead.
					} else { // I will kill that human!!
						PlayerShootSomethingHint <- player
						if player.Target.Live {
							(*gop) = append(*gop, globalOperation{
								Player:    player,
								Action:    Shoot,
								EachOther: true,
								Killed:    player.Target,
								IsResult:  true,
							})
						}
						player.Target.Kill(EatenByBeast)
					}
				} else { // My target not kill me.
					g.killPlayerNormal(player, EatenByBeast, gop)
				}
			} else { // I am not a beast
				if player.Target.Target == player { // Kill each other
					if player.Target.Status >= PlayerStatusBeast { // That is a beast...NO!
						PlayerShootSomethingHint <- player.Target
						if player.Live {
							(*gop) = append(*gop, globalOperation{
								Player:    player,
								Action:    Shoot,
								EachOther: true,
								BeKilled:  true,
								IsResult:  true,
							})
						}
						player.Kill(EatenByBeast) // I am eaten by a beast.
					} else {
						PlayerShootSomethingHint <- player        //Will hint target that player dead
						PlayerShootSomethingHint <- player.Target //Will hint player that target dead
						if player.Live && player.Target.Live {
							(*gop) = append(*gop, globalOperation{
								Player:    player,
								Action:    Shoot,
								EachOther: true,
								Killed:    player.Target,
								BeKilled:  true,
								IsResult:  true,
							})
						}
						player.Kill(Shot)
						player.Target.Kill(Shot) // All dead.
					}
				} else { // My target not kill me.
					g.killPlayerNormal(player, Shot, gop)
				}
			}
			player.User.ShootCount++
		} else if player.Target != nil && !player.Target.Live {
			(*gop) = append(*gop, globalOperation{
				Player:   player,
				Action:   Shoot,
				IsResult: true,
			})
		}
	}
}

func (g *Game) settleStageCheckTrap(gop *[]globalOperation) {
	for _, player := range g.Players {
		if player.TrapSet && player.Unioned != nil &&
			((!player.Unioned.Live && player.Unioned.KilledReason != Trapped) ||
				player.Unioned.Live) {
			if player.Status < PlayerStatusBeast {
				(*gop) = append(*gop, globalOperation{
					Player:   player,
					Action:   Trap,
					BeBeast:  true,
					IsResult: true,
				})
			}
			player.StatusChange(PlayerStatusBeast)
			player.Ununion()
		}
	}
}

func (g *Game) settleStageCheckUnion(gop *[]globalOperation) {
	for _, player := range g.Players {
		player.UnionCorrection()
	}
}

func (g *Game) settleStageExposePosition(gop *[]globalOperation) {
	for _, player := range g.Players {
		if !player.UnionValidation() && player.Live {
			if player.Status == PlayerStatusXExposed {
				(*gop) = append(*gop, globalOperation{
					Player:   player,
					Action:   Shoot,
					BeBeast:  true,
					IsResult: true,
				})
			}
			player.StatusChange()
		}
	}
}

func (g *Game) killPlayerNormal(player *Player, killedReason PlayerKilledReason, gop *[]globalOperation) {
	PlayerShootSomethingHint <- player
	if player.Target.Live {
		(*gop) = append(*gop, globalOperation{
			Player:   player,
			Action:   Shoot,
			Killed:   player.Target,
			IsResult: true,
		})
		if player.Target.Status == PlayerStatusNormal {
			player.User.GuessKillCount++
		} else if player.Target.Status == PlayerStatusXExposed {
			player.User.SniperKillCount++
		}
		player.User.KillCount++
	}
	player.Target.Kill(killedReason)
}

func (g *Game) findAbort() []*Player {
	var pl = make([]*Player, 0)
	for _, player := range g.Players {
		var found = false
		if !player.Live { // Dead no check.
			continue
		}
		for _, operation := range g.Operations {
			if operation.Player == player {
				found = true
				break
			}
		}
		if !found {
			pl = append(pl, player)
		}
	}
	return pl
}

func (g *Game) findPlayer(user *User) *Player {
	for _, player := range g.Players {
		if player.User == user {
			return player
		}
	}
	return nil
}

func (g *Game) fleeUser(user *User) {
	for i, guser := range g.Users {
		if guser.ID == user.ID {
			if i == len(g.Users)-1 { // if the last one
				g.Users = g.Users[:i] // delete the last one
			} else {
				g.Users = append(g.Users[:i], g.Users[i+1:]...)
			}
		}
	}
}

func (g *Game) makeField() ([]int, []int) {
	x := make([]int, 0)
	y := make([]int, 0)
	lenuser := len(g.Users)
	for i := 0; i < 4*lenuser*lenuser; i++ {
		xi := i / (2 * lenuser)
		yi := i % (2 * lenuser)
		pos := NewPosition(xi, yi)
		g.Positions = append(g.Positions, pos)
	}
	for i := 0; i < 2*lenuser; i++ {
		x = append(x, i)
		y = append(y, i)
	}
	return x, y
}

func (g *Game) makePlayer() {
	x, y := g.makeField()
	for _, user := range g.Users {
		rand.Seed(time.Now().Unix())
		n := rand.Intn(len(x))
		xi := x[n]
		x = append(x[:n], x[n+1:]...)

		n = rand.Intn(len(y))
		yi := y[n]
		y = append(y[:n], y[n+1:]...)

		for _, pos := range g.Positions {
			if pos.CheckPosition(xi, yi) {
				player := NewPlayer(user, pos)
				player.CurrentGamePlayersCount = len(g.Users)
				pos.BindPlayer(player) // Two-way binding
				g.Players = append(g.Players, player)
			}
		}
	}
}

func (g *Game) countDown() {
	var lock sync.RWMutex
	lock.Lock()
	defer lock.Unlock()
	if g.TimeLeft > 0 {
		g.TimeLeft--
	}
}

func (g *Game) sendPlayers() {
	PlayersHint <- g
}
