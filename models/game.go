package models

import (
	"errors"
	"math/rand"
	"time"
)

//GameStatus is
type GameStatus int

const (
	GameNotStart GameStatus = 0
	GameStart    GameStatus = 1
	GameFinished GameStatus = 2
	GameIsDay               = true
	GameIsNight             = !GameIsDay
)

//Game is
type Game struct {
	Round     int
	IsDay     bool
	Users     []*User
	Status    GameStatus
	Positions []*Position
	Players   []*Player
	TgGroup   *TgGroup
}

//NewGame is to create a new game in the group
func NewGame(tg *TgGroup) *Game {
	game := new(Game)
	game.Round = 0
	game.IsDay = GameIsDay
	game.Users = make([]*User, 0)
	game.Status = GameNotStart
	game.Positions = make([]*Position, 0)
	game.Players = make([]*Player, 0)
	game.TgGroup = tg
	return game
}

//Join is add user to game
func (g *Game) Join(user *User) {
	if g.Status == GameNotStart {
		g.Users = append(g.Users, user)
	}
}

//Flee is remove user to game or kill player in game
func (g *Game) Flee(user *User) {
	switch g.Status {
	case GameNotStart:
		g.fleeUser(user)
	case GameStart:
		p := g.findPlayer(user)
		if p != nil {
			p.Kill(Flee)
		}
	}
}

//Start is go!
func (g *Game) Start() error {
	userlen := len(g.Users)
	if userlen < 6 {
		return errors.New("Too less users")
	}
	g.makePlayer()
	return nil
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
		if guser == user {
			g.Users = append(g.Users[:i], g.Users[i+1:]...)
		}
	}
}

func (g *Game) makeField() ([]int, []int) {
	x := make([]int, 0)
	y := make([]int, 0)
	lenuser := len(g.Users)
	for i := 0; i < 4*lenuser; i++ {
		pos := NewPosition(i/(2*lenuser), i%(2*lenuser))
		g.Positions = append(g.Positions, pos)
		x = append(x, i/(2*lenuser))
		y = append(y, i%(2*lenuser))
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
			if pos.FindPosition(xi, yi) {
				player := NewPlayer(user, pos)
				pos.BindPlayer(player) // Two-way binding
				g.Players = append(g.Players, player)
			}
		}
	}
}
