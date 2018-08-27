package models

const (
	PlayerLive             = true
	PlayerDead             = false
	PlayerStatusNormal     = 0
	PlayerStatusXExposed   = 1
	PlayerStatusYExposed   = 2
	PlayerStatusWild       = PlayerStatusYExposed
	PlayerSetTrap          = true
	PlayerUnsetTrap        = false
	PlayerUnionBetrayed    = true
	PlayerUnionNotBetrayed = !PlayerUnionBetrayed
)

//PlayerKilledReason is
type PlayerKilledReason int

//List of KilledReason
const (
	None        PlayerKilledReason = -1
	Shot        PlayerKilledReason = 0
	Betrayed    PlayerKilledReason = 1
	Trapped     PlayerKilledReason = 2
	Flee        PlayerKilledReason = 3
	EatenByWild PlayerKilledReason = 4
)

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
	return player
}

//Union is used when a union request approved.
func (p *Player) Union(fromPlayer *Player) {
	p.Unioned = fromPlayer
	fromPlayer.Unioned = p
}

//Kill is to kill this player
func (p *Player) Kill(reason PlayerKilledReason) {
	p.Live = PlayerDead
	p.KilledReason = reason
}

//Shoot is
func (p *Player) Shoot(betray bool, pos *Position) *Operation {
	if betray == PlayerUnionBetrayed {
		return NewOperation(p, Shoot, p.Unioned.Position)
	}
	return NewOperation(p, Shoot, pos)
}
