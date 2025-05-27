package model

// Player holds basic profile info (loaded from auth.User).
type Player struct {
	Username    string         `json:"username"`
	Exp         int            `json:"exp"`
	Level       int            `json:"level"`
	Towers      []*Tower       `json:"towers"`
	TroopLevels map[string]int `json:"troop_levels"` // Maps troop name to level
	TowerLevels map[string]int `json:"tower_levels"` // Maps tower name to level
}

// Troop represents one unit your player can deploy.
type Troop struct {
	Name    string `json:"name"`
	HP      int    `json:"hp"`      // N/A troops (e.g. Queen) can just be zero
	ATK     int    `json:"atk"`     // attack power
	DEF     int    `json:"def"`     // defense
	Cost    int    `json:"mana"`    // alias for Mana to match API usage
	Exp     int    `json:"exp"`     // EXP reward for using/destroying?
	Special string `json:"special"` // e.g. "Heals the friendly tower with lowest HP by 300"
	Level   int    `json:"level"`   // scales stats by 10% per level
}

// Tower represents one of the three defensive buildings.
type Tower struct {
	Name  string  `json:"name"`
	HP    int     `json:"hp"`   // current hit points
	ATK   int     `json:"atk"`  // attack power (for counter‐attacks, if any)
	DEF   int     `json:"def"`  // defense
	Crit  float64 `json:"crit"` // e.g. 0.10 for 10%; 0.05 for 5%
	Exp   int     `json:"exp"`  // EXP awarded when this tower is destroyed
	Level int     `json:"level"`
}
