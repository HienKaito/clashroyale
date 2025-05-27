package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"clashroyale/internal/auth"
	"clashroyale/internal/lobby"
	"clashroyale/internal/model"
	"clashroyale/internal/upgrade"
)

type GameState struct {
	ID         string
	Players    [2]*model.Player
	Towers     [2][]*model.Tower
	Hands      [2][]*model.Troop
	Mana       map[string]int
	LastRegen  map[string]time.Time
	StartTime  time.Time
	Duration   time.Duration
	CritChance float64
	IsFinished bool
	Winner     string
	BattleLog  []string
	mu         sync.Mutex
}

var (
	mgr = struct {
		games map[string]*GameState
		mu    sync.Mutex
	}{games: make(map[string]*GameState)}

	// Global lobby manager instance
	lobbyMgr = lobby.NewManager()
)

// LoadTroops loads troop data from specs/troops.json
func LoadTroops() ([]*model.Troop, error) {
	data, err := os.ReadFile(filepath.Join("specs", "troops.json"))
	if err != nil {
		return nil, err
	}

	var troops []*model.Troop
	if err := json.Unmarshal(data, &troops); err != nil {
		return nil, err
	}
	return troops, nil
}

// LoadTowers loads tower data from specs/towers.json
func LoadTowers() ([]*model.Tower, error) {
	data, err := os.ReadFile(filepath.Join("specs", "towers.json"))
	if err != nil {
		return nil, err
	}

	var towers []*model.Tower
	if err := json.Unmarshal(data, &towers); err != nil {
		return nil, err
	}
	return towers, nil
}

// LoadPlayer loads a player from auth system
func LoadPlayer(username string) *model.Player {
	user, err := auth.LoadUser(username)
	if err != nil {
		panic(err)
	}

	// Load towers from JSON
	towers, err := LoadTowers()
	if err != nil {
		panic(err)
	}

	// Initialize level maps from user data
	troopLevels := make(map[string]int)
	towerLevels := make(map[string]int)

	// Initialize troop levels from user data
	troops, err := LoadTroops()
	if err != nil {
		panic(err)
	}
	for _, t := range troops {
		if level, ok := user.TroopLevels[t.Name]; ok {
			troopLevels[t.Name] = level
		} else {
			troopLevels[t.Name] = 1
		}
	}

	// Initialize tower levels from user data
	for _, tower := range towers {
		if level, ok := user.TowerLevels[tower.Name]; ok {
			towerLevels[tower.Name] = level
		} else {
			towerLevels[tower.Name] = 1
		}
	}

	// **Apply upgrades** to the towers themselves**
	for _, t := range towers {
		lvl := towerLevels[t.Name]
		t.HP = upgrade.CalculateUpgradeStats(t.HP, lvl)
		t.ATK = upgrade.CalculateUpgradeStats(t.ATK, lvl)
		t.DEF = upgrade.CalculateUpgradeStats(t.DEF, lvl)
	}

	return &model.Player{
		Username:    user.Username,
		Exp:         user.Exp,
		Level:       user.Level,
		Towers:      towers, // now mutated to leveled stats
		TroopLevels: troopLevels,
		TowerLevels: towerLevels,
	}
}

// cloneTowers creates a deep copy of player's towers
func cloneTowers(towers []*model.Tower) []*model.Tower {
	result := make([]*model.Tower, len(towers))
	for i, t := range towers {
		clone := *t
		result[i] = &clone
	}
	return result
}

// drawHand draws a random hand of troops for a player
func (gs *GameState) drawHand(playerIndex int) {
	troops, err := LoadTroops()
	if err != nil {
		panic(err)
	}

	// Shuffle the slice in-place
	rand.Shuffle(len(troops), func(i, j int) {
		troops[i], troops[j] = troops[j], troops[i]
	})

	// Draw up to 4 troops
	handSize := 4
	if len(troops) < handSize {
		handSize = len(troops)
	}

	hand := make([]*model.Troop, handSize)
	for i := 0; i < handSize; i++ {
		// Deep-copy each troop so we don't mutate the original spec
		tCopy := *troops[i]
		// **apply the player’s level to this clone**
		lvl := gs.Players[playerIndex].TroopLevels[tCopy.Name]
		tCopy.HP = upgrade.CalculateUpgradeStats(tCopy.HP, lvl)
		tCopy.ATK = upgrade.CalculateUpgradeStats(tCopy.ATK, lvl)
		tCopy.DEF = upgrade.CalculateUpgradeStats(tCopy.DEF, lvl)

		hand[i] = &tCopy
	}

	gs.Hands[playerIndex] = hand
}

// drawNewTroop draws a single new troop for a player
func (gs *GameState) drawNewTroop(playerIndex int) {
	troops, err := LoadTroops()
	if err != nil {
		panic(err)
	}

	// Shuffle the slice in-place
	rand.Shuffle(len(troops), func(i, j int) {
		troops[i], troops[j] = troops[j], troops[i]
	})

	// Take the first troop
	tCopy := *troops[0]
	lvl := gs.Players[playerIndex].TroopLevels[tCopy.Name]
	tCopy.HP = upgrade.CalculateUpgradeStats(tCopy.HP, lvl)
	tCopy.ATK = upgrade.CalculateUpgradeStats(tCopy.ATK, lvl)
	tCopy.DEF = upgrade.CalculateUpgradeStats(tCopy.DEF, lvl)
	gs.Hands[playerIndex] = append(gs.Hands[playerIndex], &tCopy)
}

func GetOrCreate(gameID string) *GameState {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if gs, ok := mgr.games[gameID]; ok {
		return gs
	}

	pair := lobbyMgr.GetPlayers(gameID)
	if pair[0] == "" || pair[1] == "" {
		return nil
	}

	p1 := LoadPlayer(pair[0])
	p2 := LoadPlayer(pair[1])

	gs := &GameState{
		ID:         gameID,
		Players:    [2]*model.Player{p1, p2},
		Towers:     [2][]*model.Tower{cloneTowers(p1.Towers), cloneTowers(p2.Towers)},
		Hands:      [2][]*model.Troop{make([]*model.Troop, 0), make([]*model.Troop, 0)},
		Mana:       map[string]int{p1.Username: 5, p2.Username: 5},
		LastRegen:  map[string]time.Time{p1.Username: time.Now(), p2.Username: time.Now()},
		StartTime:  time.Now(),
		Duration:   3 * time.Minute,
		CritChance: 0.1,
		IsFinished: false,
	}

	// Draw initial hands for both players
	gs.drawHand(0)
	gs.drawHand(1)

	// start the 30-second random events
	gs.startRandomEvents()

	mgr.games[gameID] = gs
	return gs
}

func (gs *GameState) RegenMana(user string) {
	now := time.Now()
	elapsed := now.Sub(gs.LastRegen[user]).Seconds()
	//regen 1 mana per sec
	if elapsed >= 1 {
		add := int(elapsed)
		gs.Mana[user] = min(10, gs.Mana[user]+add)
		gs.LastRegen[user] = now
	}
}

// AddBattleLog adds a new entry to the battle log
func (gs *GameState) AddBattleLog(entry string) {
	// No need to lock here since we're already holding the lock in Deploy
	gs.BattleLog = append(gs.BattleLog, entry)
}

// GetBattleLog returns the current battle log
func (gs *GameState) GetBattleLog() []string {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.BattleLog
}

func (gs *GameState) Deploy(username, troopName string) error {
	gs.mu.Lock()
	if gs.IsFinished {
		gs.mu.Unlock()
		return errors.New("game already finished")
	}

	//check time
	if time.Since(gs.StartTime) > gs.Duration {
		gs.mu.Unlock()
		gs.FinishGame()
		return errors.New("game time is up")
	}

	//regen mana
	gs.RegenMana(username)
	troop := gs.findAndRemoveTroop(username, troopName)
	if troop == nil {
		gs.mu.Unlock()
		return errors.New("troop not found in hand")
	}

	// Draw a new troop to replace the deployed one
	playerIndex := gs.PlayerIndex(username)
	gs.drawNewTroop(playerIndex)

	//mana cost check
	if gs.Mana[username] < troop.Cost {
		gs.mu.Unlock()
		return errors.New("not enough mana")
	}
	gs.Mana[username] -= troop.Cost

	//fin target: first alive guard > second >king
	idx := gs.PlayerIndex(username)
	enemy := 1 - idx
	var target *model.Tower

	// First try to find a Guard Tower
	for _, tw := range gs.Towers[enemy] {
		if tw.HP > 0 && tw.Name == "Guard Tower" {
			target = tw
			break
		}
	}

	// If no Guard Tower is alive, target the King Tower
	if target == nil {
		for _, tw := range gs.Towers[enemy] {
			if tw.HP > 0 && tw.Name == "King Tower" {
				target = tw
				break
			}
		}
	}

	if target == nil {
		// all dead > win
		gs.Winner = username
		gs.mu.Unlock()
		gs.FinishGame()
		return nil
	}

	// Combat loop - continue until troop is defeated or target is destroyed
	for troop.HP > 0 && target.HP > 0 {
		// Troop attacks tower
		atk := troop.ATK
		isCrit := rand.Float64() < gs.CritChance
		if isCrit {
			atk = int(float64(atk) * 1.2)
		}
		dmg := atk - target.DEF
		if dmg > 0 {
			target.HP -= dmg
			// Log the attack
			critMsg := ""
			if isCrit {
				critMsg = " (CRITICAL HIT!)"
			}
			gs.AddBattleLog(fmt.Sprintf("⚔️%s's %s attacks %s for %d damage%s", username, troop.Name, target.Name, dmg, critMsg))
			gs.AddBattleLog(fmt.Sprintf("🏰 %s now has %d HP", target.Name, target.HP))
		} else {
			gs.AddBattleLog(fmt.Sprintf("🔰%s's %s attacks %s but deals no damage (DEF too high)", username, troop.Name, target.Name))
		}

		// Check if tower is destroyed
		if target.HP <= 0 {
			gs.AddBattleLog(fmt.Sprintf("💥 %s has been destroyed!", target.Name))
			break
		}

		// Tower counter-attacks
		towerAtk := target.ATK
		towerIsCrit := rand.Float64() < target.Crit
		if towerIsCrit {
			towerAtk = int(float64(towerAtk) * 1.2)
		}
		towerDmg := towerAtk - troop.DEF
		if towerDmg > 0 {
			troop.HP -= towerDmg
			critMsg := ""
			if towerIsCrit {
				critMsg = " (CRITICAL HIT!)"
			}
			gs.AddBattleLog(fmt.Sprintf("🗡️%s counter-attacks %s for %d damage%s", target.Name, troop.Name, towerDmg, critMsg))
			gs.AddBattleLog(fmt.Sprintf("💔 %s now has %d HP", troop.Name, troop.HP))
		} else {
			gs.AddBattleLog(fmt.Sprintf("❌ %s counter-attacks %s but deals no damage (DEF too high)", target.Name, troop.Name))
		}

		// Check if troop is destroyed
		if troop.HP <= 0 {
			gs.AddBattleLog(fmt.Sprintf("☠️ %s has been defeated!", troop.Name))
			break
		}
	}

	// Check if all towers are destroyed
	allTowersDestroyed := true
	for _, tw := range gs.Towers[enemy] {
		if tw.HP > 0 {
			allTowersDestroyed = false
			break
		}
	}

	if allTowersDestroyed {
		gs.Winner = username
		gs.mu.Unlock()
		gs.FinishGame()
		return nil
	}

	if time.Since(gs.StartTime) > gs.Duration {
		gs.mu.Unlock()
		gs.FinishGame()
		return nil
	}

	gs.mu.Unlock()
	return nil
}

// find and remove
func (gs *GameState) findAndRemoveTroop(username, troopName string) *model.Troop {
	idx := gs.PlayerIndex(username)
	for i, t := range gs.Hands[idx] {
		if t.Name == troopName {
			//remove from hand
			gs.Hands[idx] = append(gs.Hands[idx][:i], gs.Hands[idx][i+1:]...)
			return t
		}
	}
	return nil
}

// player index
func (gs *GameState) PlayerIndex(username string) int {
	if gs.Players[0].Username == username {
		return 0
	}
	return 1
}

// Kick off a background ticker that applies a random event every 30s.
func (gs *GameState) startRandomEvents() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			gs.mu.Lock()
			if gs.IsFinished {
				ticker.Stop()
				gs.mu.Unlock()
				return
			}

			ev := rand.Intn(3)
			switch ev {
			case 0:
				// Heal every tower by 10 HP
				for pi := 0; pi < 2; pi++ {
					for _, tw := range gs.Towers[pi] {
						tw.HP += 10
					}
				}
				gs.AddBattleLog("🔮 Random Event: All towers healed by 10 HP")

			case 1:
				// Give every player +10 mana (cap at, say, 10)
				for user := range gs.Mana {
					gs.Mana[user] = min(10, gs.Mana[user]+10)
				}
				gs.AddBattleLog("🔮 Random Event: All players gain 10 mana")

			case 2:
				// Damage every tower by 2 HP
				for pi := 0; pi < 2; pi++ {
					for _, tw := range gs.Towers[pi] {
						tw.HP -= 2
					}
				}
				gs.AddBattleLog("🔮 Random Event: All towers take 2 damage")
			}

			gs.mu.Unlock()
		}
	}()
}

// FinishGame decides the winner and awards EXP
func (gs *GameState) FinishGame() {
	gs.mu.Lock()
	if gs.IsFinished {
		gs.mu.Unlock()
		return
	}
	gs.IsFinished = true

	//not king kill, decide tower left
	if gs.Winner == "" {
		counts := [2]int{}
		for pi := 0; pi < 2; pi++ {
			for _, tw := range gs.Towers[pi] {
				if tw.HP > 0 {
					counts[pi]++
				}
			}
		}
		switch {
		case counts[0] > counts[1]:
			gs.Winner = gs.Players[0].Username
		case counts[0] < counts[1]:
			gs.Winner = gs.Players[1].Username
		default:
			gs.Winner = "Draw"
		}
	}
	//award winner
	p1, p2 := gs.Players[0], gs.Players[1]
	switch gs.Winner {
	case p1.Username:
		p1.Exp += 30
		p2.Exp += 0
	case p2.Username:
		p1.Exp += 0
		p2.Exp += 30
	default:
		p1.Exp += 10
		p2.Exp += 10
	}

	// Add final battle log entry
	gs.AddBattleLog(fmt.Sprintf("Game Over! Winner: %s", gs.Winner))

	// Load existing user data to preserve password hash and levels
	u1, err := auth.LoadUser(p1.Username)
	if err != nil {
		panic(err)
	}
	u2, err := auth.LoadUser(p2.Username)
	if err != nil {
		panic(err)
	}

	// Update user data while preserving existing fields
	u1.Exp = p1.Exp
	u1.Level = p1.Level
	u1.TroopLevels = p1.TroopLevels
	u1.TowerLevels = p1.TowerLevels

	u2.Exp = p2.Exp
	u2.Level = p2.Level
	u2.TroopLevels = p2.TroopLevels
	u2.TowerLevels = p2.TowerLevels

	// Save updated user data
	if err := auth.SaveUser(u1); err != nil {
		panic(err)
	}
	if err := auth.SaveUser(u2); err != nil {
		panic(err)
	}

	// Remove game from lobby manager
	lobbyMgr.RemoveGame(gs.ID)
	gs.mu.Unlock()
}

// GetLobbyManager returns the global lobby manager instance
func GetLobbyManager() *lobby.Manager {
	return lobbyMgr
}
