package models

type operationResult struct {
	//Result
	Who      *Player
	Killed   string
	BeKilled bool
	BeBeast  bool
	Survive  bool // for Abort
	None     bool
	Betray   bool
}

type globalOperation struct {
	Round   int
	Player  *Player
	Action  PlayerAction
	Target  *Position // When result might nil
	Result  []operationResult
	Finally bool
}

func (g *globalOperation) AttachResult(or operationResult) {
	g.Result = append(g.Result, or)
}
