package models

import "fmt"

//PlayerAction is
type PlayerAction int

//List of player actions
const (
	Shoot  PlayerAction = 0
	Abort  PlayerAction = 1 // No action or timed out.
	Trap   PlayerAction = 2
	Betray PlayerAction = 3
)

//Operation is
type Operation struct {
	Player *Player
	Action PlayerAction
	Target *Position
}

//NewOperation is
func NewOperation(player *Player, act PlayerAction, target *Position) *Operation {
	op := new(Operation)
	op.Player = player
	op.Action = act
	op.Target = target
	return op
}

func (op *Operation) String() string {
	return fmt.Sprintf("Operation(from=%s, action=%d, target=%v)", op.Player.User.Name, op.Action, op.Target)
}
