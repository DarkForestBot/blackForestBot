package models

import "fmt"

//Position is used in game
type Position struct {
	X      int
	Y      int
	Player *Player // two-way bond
}

//NewPosition is to create new position for play
func NewPosition(x, y int) *Position {
	position := new(Position)
	position.X = x
	position.Y = y
	return position
}

//IsPosition is
func (p *Position) IsPosition(pos *Position) bool {
	return pos == p || (pos.X == p.X && pos.Y == p.Y)
}

//FindPosition is
func (p *Position) FindPosition(x, y int) bool {
	return p.X == x && p.Y == y
}

//BindPlayer is
func (p *Position) BindPlayer(player *Player) {
	p.Player = player
}

//String is
func (p *Position) String() string {
	return fmt.Sprintf("Position(%d, %d)", p.X, p.Y)
}
