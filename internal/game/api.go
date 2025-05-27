package game

import (
	"clashroyale/internal/model"
	"time"
)

// PublicState is what you'll serialize over JSON to the browser.
type PublicState struct {
	YourMana  int            `json:"yourMana"`
	YourHand  []TroopView    `json:"yourHand"`
	Towers    [2][]TowerView `json:"towers"`   // [you, opponent]
	TimeLeft  int            `json:"timeLeft"` // seconds
	Finished  bool           `json:"finished"`
	Winner    string         `json:"winner"`
	BattleLog []string       `json:"battleLog"`
}

// TroopView and TowerView are simplified versions of your full models:
type TroopView struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
	ATK  int    `json:"atk"`
	DEF  int    `json:"def"`
}

type TowerView struct {
	Name string `json:"name"`
	HP   int    `json:"hp"`
}

// Snapshot builds a PublicState for the given user
func (gs *GameState) Snapshot(user string) PublicState {
	// determine which index is "you" and "them"
	idx := gs.PlayerIndex(user)
	opp := 1 - idx

	// time left in seconds
	elapsed := time.Since(gs.StartTime)
	remaining := int(gs.Duration.Seconds() - elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// map your hand to views
	yourHand := make([]TroopView, len(gs.Hands[idx]))
	for i, t := range gs.Hands[idx] {
		yourHand[i] = TroopView{t.Name, t.Cost, t.ATK, t.DEF}
	}

	// map towers
	toViews := func(towers []*model.Tower) []TowerView {
		vs := make([]TowerView, len(towers))
		for i, tw := range towers {
			vs[i] = TowerView{tw.Name, tw.HP}
		}
		return vs
	}

	return PublicState{
		YourMana:  gs.Mana[user],
		YourHand:  yourHand,
		Towers:    [2][]TowerView{toViews(gs.Towers[idx]), toViews(gs.Towers[opp])},
		TimeLeft:  remaining,
		Finished:  gs.IsFinished,
		Winner:    gs.Winner,
		BattleLog: gs.GetBattleLog(),
	}
}
