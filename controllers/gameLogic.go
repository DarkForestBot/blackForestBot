package controllers

import "git.wetofu.top/tonychee7000/blackForestBot/models"

func gameLogic(game *models.Game) {
	// Stage I: tag the target and beast the player abort
	for _, opeartion := range game.Operations {
		switch opeartion.Action {
		case models.Shoot: // Betray is special shoot
			if opeartion.Target != nil && opeartion.Target.Player != nil {
				opeartion.Player.Target = opeartion.Target.Player
				if opeartion.Target.Player.KilledBy == nil {
					opeartion.Target.Player.KilledBy = opeartion.Player
				}
			}
		case models.Abort:
			opeartion.Player.StatusChange(models.PlayerStatusBeast)
		}
	}
	// Stage II: check betray.
	for _, player := range game.Players {
		if player.Live == models.PlayerDead || player.Unioned == nil ||
			player.Unioned.Unioned == nil || player.Unioned.Unioned != player { // Here must some mistakes.
			continue
		}
		if player.Target != nil && player.Target == player.Unioned { // Betray
			player.User.BetrayCount++
			if player.KilledBy != nil && player.KilledBy == player.Unioned { // Betray each other
				player.Unioned.User.BetrayCount++
				player.StatusChange(models.PlayerStatusBeast)
				player.Unioned.StatusChange(models.PlayerStatusBeast)
				player.Target = nil
				player.Unioned.Target = nil
				player.Ununion() // Union broken.
			} else { // I betrayed my union
				if player.Target.TrapSet {
					player.Kill(models.Trapped) // Oops! I was trapped!
				} else {
					player.Target.Kill(models.Betrayed)
					player.Target = nil
				}
				player.Ununion() // Union broken.
			}
		}
	}
	// Stage III: check who surely dead.
	for _, player := range game.Players {
		if player.Live == models.PlayerDead {
			continue
		}
		if player.Target != nil && player.Target.Live { // I want to kill some one.
			if player.Status >= models.PlayerStatusBeast { // I am a beast
				if player.KilledBy != nil && player.Target == player.KilledBy { // Kill each other
					if player.Target.Status >= models.PlayerStatusBeast { // He is a beast also...NO!!
						player.Target.Kill(models.BeastKill)
						player.Kill(models.BeastKill) // All dead.
					} else { // I will kill that human!!
						player.Target.Kill(models.EatenByBeast)
					}
				} else { // My target not kill me.
					player.Target.Kill(models.EatenByBeast)
					player.User.KillCount++
				}
			} else { // I am not a beast
				if player.KilledBy != nil && player.Target == player.KilledBy { // Kill each other
					if player.Target.Status >= models.PlayerStatusBeast { // He is a beast...NO!
						player.Kill(models.EatenByBeast) // I am eaten by a beast.
					} else {
						player.Kill(models.Shot)
						player.Target.Kill(models.Shot) // All dead.
					}
				} else { // My target not kill me.
					player.Target.Kill(models.Shot)
				}
			}
			player.User.ShootCount++
		}
	}
	// Stage IV: check trap
	for _, player := range game.Players {
		if player.TrapSet && player.Unioned != nil &&
			((!player.Unioned.Live && player.Unioned.KilledReason != models.Trapped) ||
				player.Unioned.Live) {
			player.StatusChange(models.PlayerStatusBeast)
			player.Ununion()
		}
	}
	// Stage V: check union
	for _, player := range game.Players {
		if !player.Live { // Dead man no union
			player.Ununion()
		}
	}
	// Stage VI: expose position
	for _, player := range game.Players {
		if player.Unioned == nil {
			player.StatusChange()
		}
	}
	// Stage 0: Reset some status
	for _, player := range game.Players {
		game.Operations = make([]*models.Operation, 0)
		player.ActionClear()
		player.User.Update()
	}
}
