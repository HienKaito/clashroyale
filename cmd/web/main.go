package main

import (
	"clashroyale/internal/auth"
	"clashroyale/internal/game"
	"clashroyale/internal/model"
	"clashroyale/internal/upgrade"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/static", "./templates/static")
	r.LoadHTMLGlob("cmd/web/templates/*")

	// session store (32-byte secret)
	store := cookie.NewStore([]byte("a-very-secret-key-1234567890abcdef"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(24 * time.Hour / time.Second),
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("tcrsess", store))

	r.GET("/register", showRegister)
	r.POST("/register", doRegister)

	r.GET("/login", showLogin)
	r.POST("/login", doLogin)

	r.GET("/dashboard", authRequired(), dashboard)

	r.GET("/logout", func(c *gin.Context) {
		sess := sessions.Default(c)
		sess.Clear()
		sess.Save()
		c.Redirect(http.StatusSeeOther, "/login")
	})

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/login")
	})

	r.GET("/lobby", authRequired(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "lobby.html", gin.H{
			"QueueLen": game.GetLobbyManager().QueueLength(),
		})
	})

	r.POST("/lobby/join", authRequired(), func(c *gin.Context) {
		user := sessions.Default(c).Get("user").(string)
		if gameID := game.GetLobbyManager().Join(user); gameID != "" {
			c.Redirect(http.StatusSeeOther, "/game/"+gameID)
			return
		}
		c.Redirect(http.StatusSeeOther, "/lobby/wait")
	})

	r.GET("/lobby/wait", authRequired(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "wait.html", nil)
	})

	r.GET("/lobby/status", authRequired(), func(c *gin.Context) {
		user := sessions.Default(c).Get("user").(string)
		if gameID := game.GetLobbyManager().GetGame(user); gameID != "" {
			c.JSON(http.StatusOK, gin.H{"gameID": gameID})
		} else {
			c.JSON(http.StatusOK, gin.H{"gameID": ""})
		}
	})

	r.GET("/game/:gameID", authRequired(), func(c *gin.Context) {
		id := c.Param("gameID")
		players := game.GetLobbyManager().GetPlayers(id)
		c.HTML(http.StatusOK, "game.html", gin.H{
			"GameID":  id,
			"Players": players,
		})
	})

	r.GET("/game/:gameID/state", authRequired(), func(c *gin.Context) {
		gs := game.GetOrCreate(c.Param("gameID"))
		user := sessions.Default(c).Get("user").(string)

		gs.RegenMana(user)
		if time.Since(gs.StartTime) > gs.Duration && !gs.IsFinished {
			gs.FinishGame()
		}
		c.JSON(http.StatusOK, gs.Snapshot(user))
	})

	r.POST("/game/:gameID/deploy", authRequired(), func(c *gin.Context) {
		tr := c.PostForm("troop")
		user := sessions.Default(c).Get("user").(string)
		err := game.GetOrCreate(c.Param("gameID")).Deploy(user, tr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": true})
		}
	})

	r.POST("/upgrade/troop", authRequired(), upgradeTroop)
	r.POST("/upgrade/tower", authRequired(), upgradeTower)

	r.Run(":8080")
}

// Middleware to require login
func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		user := sess.Get("user")
		if user == nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func showRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func doRegister(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if _, err := auth.Register(username, password); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"Error": err.Error()})
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
}

func showLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func doLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := auth.Authenticate(username, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": err.Error()})
		return
	}

	sess := sessions.Default(c)
	sess.Set("user", user.Username)
	sess.Save()

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func dashboard(c *gin.Context) {
	username := sessions.Default(c).Get("user").(string)
	player := game.LoadPlayer(username)

	// Load troops for upgrade display
	troops, err := game.LoadTroops()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load troops"})
		return
	}

	// Prepare troop data for display
	var troopData []gin.H
	for _, base := range troops {
		lvl := player.TroopLevels[base.Name]
		canUpgrade, cost := upgrade.CanUpgradeTroop(player, base)

		// **re‐calculate** stats from the ORIGINAL base JSON values
		hp := upgrade.CalculateUpgradeStats(base.HP, lvl)
		atk := upgrade.CalculateUpgradeStats(base.ATK, lvl)
		def := upgrade.CalculateUpgradeStats(base.DEF, lvl)

		troopData = append(troopData, gin.H{
			"Name":        base.Name,
			"HP":          hp,
			"ATK":         atk,
			"DEF":         def,
			"Level":       lvl,
			"UpgradeCost": cost,
			"CanUpgrade":  canUpgrade,
		})
	}

	// Prepare tower data for display (already upgraded in LoadPlayer)
	var towerData []gin.H
	for _, tower := range player.Towers {
		lvl := player.TowerLevels[tower.Name]
		canUpgrade, cost := upgrade.CanUpgradeTower(player, tower)

		towerData = append(towerData, gin.H{
			"Name":        tower.Name,
			"HP":          tower.HP,  // use the mutated HP
			"ATK":         tower.ATK, // use the mutated ATK
			"DEF":         tower.DEF, // use the mutated DEF
			"Level":       lvl,
			"UpgradeCost": cost,
			"CanUpgrade":  canUpgrade,
		})
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Username": username,
		"Exp":      player.Exp,
		"Level":    player.Level,
		"Troops":   troopData,
		"Towers":   towerData,
	})
}

// Add upgrade endpoints
func upgradeTroop(c *gin.Context) {
	username := sessions.Default(c).Get("user").(string)
	player := game.LoadPlayer(username)
	troopName := c.PostForm("name")

	// Find the troop
	troops, err := game.LoadTroops()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load troops"})
		return
	}

	var targetTroop *model.Troop
	for _, troop := range troops {
		if troop.Name == troopName {
			targetTroop = troop
			break
		}
	}

	if targetTroop == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Troop not found"})
		return
	}

	// Attempt upgrade
	success, err := upgrade.UpgradeTroop(player, targetTroop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough EXP or cannot upgrade"})
		return
	}

	// Load existing user data to preserve password hash
	user, err := auth.LoadUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
		return
	}

	// Update user data while preserving password hash
	user.Exp = player.Exp
	user.Level = player.Level
	user.TroopLevels = player.TroopLevels
	user.TowerLevels = player.TowerLevels

	// Save updated player data
	if err := auth.SaveUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save player data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func upgradeTower(c *gin.Context) {
	username := sessions.Default(c).Get("user").(string)
	player := game.LoadPlayer(username)
	towerName := c.PostForm("name")

	// Find the tower
	var targetTower *model.Tower
	for _, tower := range player.Towers {
		if tower.Name == towerName {
			targetTower = tower
			break
		}
	}

	if targetTower == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tower not found"})
		return
	}

	// Attempt upgrade
	success, err := upgrade.UpgradeTower(player, targetTower)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough EXP or cannot upgrade"})
		return
	}

	// Load existing user data to preserve password hash
	user, err := auth.LoadUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
		return
	}

	// Update user data while preserving password hash
	user.Exp = player.Exp
	user.Level = player.Level
	user.TroopLevels = player.TroopLevels
	user.TowerLevels = player.TowerLevels

	// Save updated player data
	if err := auth.SaveUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save player data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
