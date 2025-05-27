# Clash Royale Go Server

A Go-based web server that brings a Clash Royale–style experience to your browser. Players can register, log in, upgrade troops and towers, join a matchmaking lobby, and battle in real time—all powered by Go, Gin, and simple JSON specs.

---

## 🚀 Features

- **User Authentication**: Register & Login with session storage  
- **Dashboard**: View EXP, Player Level, Troop & Tower stats  
- **Upgrade System**: Spend EXP to upgrade individual troops and towers  
- **Matchmaking Lobby**: Join a queue and wait for an opponent  
- **Real-Time Battles**: Deploy troops, towers auto-attack, battle log updates  
- **Random Events**: Every 30 seconds triggers one of three global events (heal towers, mana boost, tower damage)  

---

## 🗂 Project Structure

```
clashroyale/
├── cmd/
│   └── web/
│       ├── main.go             # HTTP server entrypoint (Gin)
│       └── templates/          # HTML templates
│           ├── login.html
│           ├── register.html
│           ├── dashboard.html
│           ├── lobby.html
│           ├── wait.html
│           └── game.html
├── internal/
│   ├── auth/                   # User registration & authentication
│   ├── game/                   # Matchmaking and game logic
│   ├── model/                  # Data models (Player, Troop, Tower)
│   └── upgrade/                # Upgrade cost/stat calculations
├── specs/
│   ├── troops.json             # Base stats for all troops
│   └── towers.json             # Base stats for all towers
├── static/
│   └── images/                 # Backgrounds, icons, etc.
├── go.mod
├── go.sum
└── README.md
```

---

## ⚙️ Getting Started

### Prerequisites

- **Go 1.18+**  
- Git

### Installation

```bash
# 1. Clone the repo
git clone https://github.com/HienKaito/clashroyale.git
cd clashroyale

# 2. Download dependencies
go mod download
```

### Running

```bash
# From project root
go run cmd/web/main.go
```

Visit [http://localhost:8080](http://localhost:8080) in your browser.

---

## 🕹️ Usage

1. **Register** a new account at `/register`  
2. **Log in** at `/login`  
3. **Dashboard**:  
   - Spend EXP to upgrade troops & towers  
   - View your current level, EXP, and unit stats  
4. **Join Lobby**: click “Go to Lobby” and wait for an opponent  
5. **Battle**:  
   - Deploy troops from your hand (costs mana)  
   - Watch your towers auto-attack  
   - See the battle log update in real time  
6. **Random Events**:  
   - Every 30 s, one of the following triggers globally:  
     - Heal all towers by 10 HP  
     - +10 mana to both players  
     - Deal 2 HP damage to all towers  

---

## 🤝 Contributing

1. Fork the repo  
2. Create a branch: `git checkout -b feature/YourFeature`  
3. Commit your changes: `git commit -m "Add YourFeature"`  
4. Push to your branch: `git push origin feature/YourFeature`  
5. Open a Pull Request

Please follow the existing code style and write tests for new functionality when possible.

---

## 📄 License

This project is licensed under the MIT License. Feel free to use, modify, and distribute!  
