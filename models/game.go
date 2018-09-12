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
	GlobalOperations [][]*globalOperation
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
	if g.Status != GameNotStart {
		return nil
	}

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
			&globalOperation{
				Round:   g.Round,
				Finally: true,
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
			&globalOperation{
				Round:   g.Round,
				Player:  g.Winner,
				Finally: true,
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
	// Stage I: tag the target and beast the player abort
	g.settleStageTag()
	// Stage II: check betray.
	g.settleStageCheckBetry()
	// Stage III: check who surely dead.
	g.settleStageCheckDeath()
	// Stage IV: check trap
	g.settleStageCheckTrap()
	// Stage V: check union
	g.settleStageCheckUnion()
	// Stage VI: expose position
	g.settleStageExposePosition()
	// Stage O: Reset some status
	for i, player := range g.Players {
		if !player.Live {
			if i < len(g.Players)-1 {
				// It seems sort...
				g.Players = append(g.Players[:i], g.Players[i+1:]...)
				g.Players = append(g.Players, player)
			}
		}
		player.ActionClear()
		player.User.Update()
	}
	g.Operations = make([]*Operation, 0)
}

func (g *Game) settleStageTag() {
	var operations = make([]*globalOperation, 0)
	for _, operation := range g.Operations {
		operations = append(operations, &globalOperation{
			Round:  g.Round,
			Player: operation.Player,
			Action: operation.Action,
			Target: operation.Target,
			Result: make([]operationResult, 0),
		})
	}
	g.GlobalOperations = append(g.GlobalOperations, operations)

	for _, operation := range g.Operations {
		switch operation.Action {
		case Shoot: // Betray is special shoot
			if operation.Target != nil && operation.Target.Player != nil {
				operation.Player.ShootNoneStreak = 0
				operation.Player.Target = operation.Target.Player
			} else {
				operation.Player.User.ShootCount++
				operation.Player.ShootNoneStreak++
				PlayerShootNothingHint <- operation.Player
				op := g.findGlobalOperation(g.Round, operation.Player)
				if op != nil {
					op.AttachResult(operationResult{
						Who:  operation.Player,
						None: true,
					})
				}
			}
		case Abort:
			rand.Seed(time.Now().Unix())
			operation.Player.ShootNoneStreak = 0
			if operation.Player.Status < PlayerStatusBeast {
				fate := rand.Intn(4) // 1/n rate
				if fate == 0 {
					if operation.Player.Status < PlayerStatusBeast {
						op := g.findGlobalOperation(g.Round, operation.Player)
						if op != nil {
							op.AttachResult(operationResult{
								Who:     operation.Player,
								BeBeast: true,
							})
						}
					}
					operation.Player.StatusChange(PlayerStatusBeast)
				} else if fate == 1 {
					if operation.Player.Live {
						op := g.findGlobalOperation(g.Round, operation.Player)
						if op != nil {
							op.AttachResult(operationResult{
								Who:      operation.Player,
								BeKilled: true,
							})
						}
					}
					operation.Player.Kill(Flee)
				} else {
					PlayerSurvivedAtNightHint <- operation.Player
					op := g.findGlobalOperation(g.Round, operation.Player)
					if op != nil {
						op.AttachResult(operationResult{
							Who:     operation.Player,
							Survive: true,
						})
					}
				}
			} else {
				if rand.Intn(3) == 0 { // 1/n
					if operation.Player.Live {
						op := g.findGlobalOperation(g.Round, operation.Player)
						if op != nil {
							op.AttachResult(operationResult{
								Who:      operation.Player,
								BeKilled: true,
							})
						}
					}
					operation.Player.Kill(Flee)
				} else {
					PlayerSurvivedAtNightHint <- operation.Player
					op := g.findGlobalOperation(g.Round, operation.Player)
					if op != nil {
						op.AttachResult(operationResult{
							Who:     operation.Player,
							Survive: true,
						})
					}
				}
			}
		}
	}
}

func (g *Game) settleStageCheckBetry() {
	for _, player := range g.Players {
		if !player.UnionValidation() { // Here must some mistakes.
			player.Ununion()
			continue
		}
		if player.Target != nil && player.Target == player.Unioned { // Betray
			player.User.BetrayCount++
			if player.Unioned.Target == player { // Betray each other
				player.Unioned.User.BetrayCount++
				op := g.findGlobalOperation(g.Round, player)
				if op != nil {
					if player.Status < PlayerStatusBeast {
						op.AttachResult(operationResult{
							Who:     player,
							BeBeast: true,
						})
					}
					if player.Unioned.Status < PlayerStatusBeast {
						op.AttachResult(operationResult{
							Who:     player.Unioned,
							BeBeast: true,
						})
					}
				}
				player.StatusChange(PlayerStatusBeast)
				player.Unioned.StatusChange(PlayerStatusBeast)
				player.Target = nil
				player.Unioned.Target = nil
				player.Ununion() // Union broken.
			} else { // I betrayed my union
				if player.Target.TrapSet {
					if player.Live {
						op := g.findGlobalOperation(g.Round, player)
						if op != nil {
							op.AttachResult(operationResult{
								Who:      player,
								BeKilled: true,
							})
						}
					}
					player.Kill(Trapped) // Oops! I was trapped!
					player.User.KilledByTrapCount++
					player.Target.ShootNoneStreak = 0
				} else {
					if player.Target.Live {
						op := g.findGlobalOperation(g.Round, player)
						if op != nil {
							op.AttachResult(operationResult{
								Who:    player,
								Killed: player.Target.User.Name,
								Betray: true,
							})
						}
					}
					player.Target.Kill(Betrayed)
					player.Target = nil
				}
			}
		}
	}
}

func (g *Game) settleStageCheckDeath() {
	for _, player := range g.Players {
		if !player.Live && player.KilledReason == Trapped { // I was trapped, and dead now.
			continue
		}
		if player.Target != nil && player.Target.Live { // I want to kill some one.
			if player.Status >= PlayerStatusBeast { // I am a beast
				if player.Target.Target == player { // Kill each other
					if player.Target.Status >= PlayerStatusBeast { // That is a beast also...NO!!
						PlayerShootSomethingHint <- player.Target //Will hint target that player dead
						PlayerShootSomethingHint <- player        //Will hint player that target dead
						if player.Target.Live && player.Live {
							op := g.findGlobalOperation(g.Round, player)
							if op != nil {
								op.AttachResult(operationResult{
									Who:      player,
									Killed:   player.Target.User.Name,
									BeKilled: true,
								})
								op.AttachResult(operationResult{
									Who:      player.Target,
									Killed:   player.User.Name,
									BeKilled: true,
								})
							}
						}
						player.Target.Kill(BeastKill)
						player.Kill(BeastKill) // All dead.
					} else { // I will kill that human!!
						PlayerShootSomethingHint <- player
						if player.Target.Live {
							op := g.findGlobalOperation(g.Round, player)
							if op != nil {
								op.AttachResult(operationResult{
									Who:    player,
									Killed: player.Target.User.Name,
								})
							}
						}
						player.Target.Kill(EatenByBeast)
					}
				} else { // My target not kill me.
					g.killPlayerNormal(player, EatenByBeast)
				}
			} else { // I am not a beast
				if player.Target.Target == player { // Kill each other
					if player.Target.Status >= PlayerStatusBeast { // That is a beast...NO!
						PlayerShootSomethingHint <- player.Target
						if player.Live {
							op := g.findGlobalOperation(g.Round, player)
							if op != nil {
								op.AttachResult(operationResult{
									Who:      player,
									BeKilled: true,
								})
							}
						}
						player.Kill(EatenByBeast) // I am eaten by a beast.
					} else {
						PlayerShootSomethingHint <- player        //Will hint target that player dead
						PlayerShootSomethingHint <- player.Target //Will hint player that target dead
						if player.Live && player.Target.Live {
							op := g.findGlobalOperation(g.Round, player)
							if op != nil {
								op.AttachResult(operationResult{
									Who:      player,
									BeKilled: true,
									Killed:   player.Target.User.Name,
								})
								op.AttachResult(operationResult{
									Who:      player.Target,
									Killed:   player.User.Name,
									BeKilled: true,
								})
							}
						}
						player.Kill(Shot)
						player.Target.Kill(Shot) // All dead.
					}
				} else { // My target not kill me.
					g.killPlayerNormal(player, Shot)
				}
			}
			player.User.ShootCount++
		} else if player.Target != nil && !player.Target.Live {
			if player.Live {
				op := g.findGlobalOperation(g.Round, player)
				if op != nil {
					op.AttachResult(operationResult{
						Who:  player,
						None: true,
					})
				}
			}
		}
	}
}

func (g *Game) settleStageCheckTrap() {
	for _, player := range g.Players {
		if player.TrapSet && player.Unioned != nil &&
			((!player.Unioned.Live && player.Unioned.KilledReason != Trapped) ||
				player.Unioned.Live) {
			if player.Status < PlayerStatusBeast {
				op := g.findGlobalOperation(g.Round, player)
				if op != nil {
					op.AttachResult(operationResult{
						Who:     player,
						BeBeast: true,
					})
				}
			}
			player.StatusChange(PlayerStatusBeast)
			player.Ununion()
		}
	}
}

func (g *Game) settleStageCheckUnion() {
	for _, player := range g.Players {
		player.UnionCorrection()
	}
}

func (g *Game) settleStageExposePosition() {
	for _, player := range g.Players {
		if !player.UnionValidation() && player.Live {
			if player.Status == PlayerStatusXExposed || player.Status == PlayerStatusYExposed {
				op := g.findGlobalOperation(g.Round, player)
				if op != nil {
					op.AttachResult(operationResult{
						Who:     player,
						BeBeast: true,
					})
				}
			}
			player.StatusChange()
		}
	}
}

func (g *Game) killPlayerNormal(player *Player, killedReason PlayerKilledReason) {
	PlayerShootSomethingHint <- player
	if player.Target.Live {
		op := g.findGlobalOperation(g.Round, player)
		if op != nil {
			op.AttachResult(operationResult{
				Who:    player,
				Killed: player.Target.User.Name,
			})
		}
		if player.Target.Status == PlayerStatusNormal {
			player.User.GuessKillCount++
		} else if player.Target.Status == PlayerStatusXExposed ||
			player.Target.Status == PlayerStatusYExposed {
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

func (g *Game) findGlobalOperation(round int, player *Player) *globalOperation {
	defer func() { recover() }()
	for _, op := range g.GlobalOperations[round-1] {
		if op.Player == player {
			return op
		}
	}
	return nil
}
