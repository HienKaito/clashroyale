package upgrade

import (
	"clashroyale/internal/model"
	"math"
)

// CalculateUpgradeCost calculates the EXP cost for upgrading using
// the unit's own baseCost, increasing by 10% per level (rounded down).
func CalculateUpgradeCost(baseCost, currentLevel int) int {
	return int(float64(baseCost) * math.Pow(1.1, float64(currentLevel)))
}

// CalculateUpgradeStats calculates the new stats after an upgrade
// using the base stat and current level
func CalculateUpgradeStats(baseStat, level int) int {
	val := float64(baseStat) * math.Pow(1.1, float64(level))
	return int(math.Ceil(val))
}

// CanUpgradeTroop checks if a player can upgrade a specific troop
// using that troop's own Exp value as the base cost.
func CanUpgradeTroop(player *model.Player, troop *model.Troop) (bool, int) {
	lvl := player.TroopLevels[troop.Name]
	cost := CalculateUpgradeCost(troop.Exp, lvl)
	return player.Exp >= cost, cost
}

// CanUpgradeTower checks if a player can upgrade a specific tower
// using that tower's own Exp value as the base cost.
func CanUpgradeTower(player *model.Player, tower *model.Tower) (bool, int) {
	lvl := player.TowerLevels[tower.Name]
	cost := CalculateUpgradeCost(tower.Exp, lvl)
	return player.Exp >= cost, cost
}

// UpgradeTroop upgrades a troop (deducts EXP, bumps its level, recalculates stats).
func UpgradeTroop(player *model.Player, troop *model.Troop) (bool, error) {
	can, cost := CanUpgradeTroop(player, troop)
	if !can {
		return false, nil
	}

	player.Exp -= cost
	newLevel := player.TroopLevels[troop.Name] + 1
	player.TroopLevels[troop.Name] = newLevel

	return true, nil
}

// UpgradeTower upgrades a tower similarly.
func UpgradeTower(player *model.Player, tower *model.Tower) (bool, error) {
	can, cost := CanUpgradeTower(player, tower)
	if !can {
		return false, nil
	}

	player.Exp -= cost
	newLevel := player.TowerLevels[tower.Name] + 1
	player.TowerLevels[tower.Name] = newLevel

	return true, nil
}
