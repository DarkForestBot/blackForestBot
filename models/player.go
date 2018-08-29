package models

import (
	"fmt"
	"log"
)

//List of some consts
const (
	PlayerLive             = true
	PlayerDead             = false
	PlayerStatusNormal     = 0
	PlayerStatusXExposed   = 1
	PlayerStatusYExposed   = 2
	PlayerStatusBeast      = PlayerStatusYExposed
	PlayerSetTrap          = true
	PlayerUnsetTrap        = false
	PlayerUnionBetrayed    = true
	PlayerUnionNotBetrayed = !PlayerUnionBetrayed
	PlayerShootNone        = -1
)

//PlayerKilledReason is
type PlayerKilledReason int

//List of KilledReason
const (
	None         PlayerKilledReason = -1
	Shot         PlayerKilledReason = 0
	Betrayed     PlayerKilledReason = 1
	Trapped      PlayerKilledReason = 2
	Flee         PlayerKilledReason = 3
	EatenByBeast PlayerKilledReason = 4
	BeastKill    PlayerKilledReason = 5 // Beasts killed each other
)

var reasonStrings = []string{
	"shot",
	"betrayed",
	"trapped",
	"eaten by a beast",
	"beasts killed each other",
}

//Player is used in redis
type Player struct {
	User         *User
	Live         bool
	KilledReason PlayerKilledReason
	Position     *Position // two-way bond
	Unioned      *Player
	Grouped      *Group
	Job          *Job //TODO: soon tm
	Status       int
	TrapSet      bool
	ShootX       int // Every round clear
	ShootY       int // Every round clear
	UnionReqs    []int
	KilledBy     *Player // Will be killed by whom
	Target       *Player // Will kill whom
}

//NewPlayer is called when start a game
func NewPlayer(user *User, position *Position) *Player {
	player := new(Player)
	player.User = user
	player.Live = PlayerLive
	player.KilledReason = None
	player.Position = position
	player.Status = PlayerStatusNormal
	player.TrapSet = PlayerUnsetTrap
	player.UnionReqs = make([]int, 0)
	user.GamesJoined++
	user.Update()
	return player
}

//Union is used when a union request approved.
func (p *Player) Union(fromPlayer *Player) {
	p.Unioned = fromPlayer
	fromPlayer.Unioned = p
}

//Ununion is called when betray or one man dead
func (p *Player) Ununion() {
	defer func() {
		recover() // might panic
	}()
	p.Unioned.Unioned = nil
	p.Unioned = nil
}

//Kill is to kill this player
func (p *Player) Kill(reason PlayerKilledReason) {
	if !p.Live {
		return
	}
	p.Live = PlayerDead
	p.KilledReason = reason
	log.Printf("Player(%s) Killed for reason `%s`.\n", p.User.Name, reasonStrings[reason])
}

//Shoot is
func (p *Player) Shoot(betray bool, pos *Position) *Operation {
	if betray == PlayerUnionBetrayed {
		return NewOperation(p, Shoot, p.Unioned.Position)
	}
	return NewOperation(p, Shoot, pos)
}

// ActionClear is
func (p *Player) ActionClear() {
	p.ShootX = PlayerShootNone
	p.ShootY = PlayerShootNone
	p.TrapSet = false
	p.Target = nil
	if p.Status == PlayerStatusBeast {
		p.KilledBy = nil
	}
}

// StatusChange is
func (p *Player) StatusChange(stage ...int) {
	if stage != nil && len(stage) > 0 {
		p.Status = stage[0]
	} else {
		p.Status++
	}
	if p.Status >= PlayerStatusBeast {
		p.Status = PlayerStatusBeast
	}
}

// GetPositionString is a backup method
func (p *Player) GetPositionString() string {
	switch p.Status {
	case PlayerStatusNormal:
		return "(?, ?)"
	case PlayerStatusXExposed:
		return fmt.Sprintf("(%d, ?)", p.Position.X)
	case PlayerStatusBeast:
		return fmt.Sprintf("(%d, %d)", p.Position.X, p.Position.Y)
	}
	return ""
}
