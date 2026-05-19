package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_UpdateHeartbeat(t *testing.T) {
	registry := NewRegistry()
	agentID := "agent-123"

	// Initial update
	registry.UpdateHeartbeat(agentID)

	// Check if present
	registry.mu.RLock()
	lastSeen, ok := registry.agents[agentID]
	registry.mu.RUnlock()

	assert.True(t, ok)
	assert.WithinDuration(t, time.Now(), lastSeen, time.Second)
}

func TestRegistry_GetActiveAgents(t *testing.T) {
	registry := NewRegistry()
	agent1 := "agent-1"
	agent2 := "agent-2"
	agent3 := "agent-3"

	// Mock current time
	now := time.Now()

	registry.agents[agent1] = now // Just updated
	registry.agents[agent2] = now.Add(-5 * time.Second) // Updated 5 seconds ago
	registry.agents[agent3] = now.Add(-15 * time.Second) // Updated 15 seconds ago

	activeAgents := registry.GetActiveAgents(10 * time.Second)

	assert.Len(t, activeAgents, 2)
	assert.Contains(t, activeAgents, agent1)
	assert.Contains(t, activeAgents, agent2)
	assert.NotContains(t, activeAgents, agent3)
}
