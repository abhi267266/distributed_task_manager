package server

import (
	"sync"
	"time"
)

// Registry maintains an in-memory list of online agents.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]time.Time
}

// NewRegistry creates a new agent Registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]time.Time),
	}
}

// UpdateHeartbeat updates the last seen timestamp for a given agent.
func (r *Registry) UpdateHeartbeat(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agentID] = time.Now()
}

// GetActiveAgents returns a list of agent IDs that have sent a heartbeat within the specified timeout.
func (r *Registry) GetActiveAgents(timeout time.Duration) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []string
	now := time.Now()
	for id, lastSeen := range r.agents {
		if now.Sub(lastSeen) <= timeout {
			active = append(active, id)
		}
	}
	return active
}
