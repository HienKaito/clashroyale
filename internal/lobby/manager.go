package lobby

import (
	"sync"

	"github.com/google/uuid"
)

type Manager struct {
	queue []string
	games map[string][2]string
	mu    sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		queue: make([]string, 0),
		games: make(map[string][2]string),
	}
}

func (m *Manager) Join(username string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	//already in game
	for id, ps := range m.games {
		if ps[0] == username || ps[1] == username {
			return id
		}
	}

	//already in queue
	for _, u := range m.queue {
		if u == username {
			return ""
		}
	}

	//enqueue
	m.queue = append(m.queue, username)

	// if 2+, pair them
	if len(m.queue) >= 2 {
		p1, p2 := m.queue[0], m.queue[1]
		m.queue = m.queue[2:]
		id := uuid.NewString()
		m.games[id] = [2]string{p1, p2}
		return id
	}
	return ""
}

// GetGame
func (m *Manager) GetGame(username string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, ps := range m.games {
		if ps[0] == username || ps[1] == username {
			return id
		}
	}
	return ""
}

// GetPlayers
func (m *Manager) GetPlayers(gameID string) [2]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.games[gameID]
}

// QueueLength
func (m *Manager) QueueLength() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.queue)
}

// RemoveGame removes a game from the active games list
func (m *Manager) RemoveGame(gameID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.games, gameID)
}
