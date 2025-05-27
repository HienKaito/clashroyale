package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username     string         `json:"username"`
	PasswordHash string         `json:"password_hash"`
	Exp          int            `json:"exp"`
	Level        int            `json:"level"`
	TroopLevels  map[string]int `json:"troop_levels"` // Maps troop name to level
	TowerLevels  map[string]int `json:"tower_levels"` // Maps tower name to level
}

var dataDir = filepath.Join("data", "players")

func init() {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(fmt.Sprintf("Cannot create data directory: %v", err))
	}
}

// Hashpass
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// check password
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Load user
func LoadUser(username string) (*User, error) {
	path := filepath.Join(dataDir, username+".json")
	f, err := os.Open(path)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf("user %s not found", username)
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	var u User
	if err := json.NewDecoder(f).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Save user
func SaveUser(u *User) error {
	path := filepath.Join(dataDir, u.Username+".json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(u)
}

// reegister
func Register(username, password string) (*User, error) {
	if _, err := LoadUser(username); err == nil {
		return nil, errors.New("user already exists")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &User{
		Username:     username,
		PasswordHash: hash,
		Exp:          0,
		Level:        0,
		TroopLevels:  make(map[string]int),
		TowerLevels:  make(map[string]int),
	}
	if err := SaveUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

// Authen check
func Authenticate(username, password string) (*User, error) {
	u, err := LoadUser(username)
	if err != nil {
		return nil, err
	}
	if err := CheckPassword(password, u.PasswordHash); err != nil {
		return nil, errors.New("invalid password")
	}
	return u, nil
}
