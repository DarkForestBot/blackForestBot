package controllers

import (
	"git.wetofu.top/tonychee7000/blackForestBot/models"
)

var gameList map[int64]*models.Game

func init() {
	gameList = make(map[int64]*models.Game)
}
