package agent

import (
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

// LoadOrGenerateID checks if the identity file exists.
// If it does, it reads and returns the UUID.
// If it doesn't, it generates a new UUID, saves it to the file, and returns it.
func LoadOrGenerateID(path string) (string, error) {
	// Check if file exists
	content, err := os.ReadFile(path)
	if err == nil {
		id := strings.TrimSpace(string(content))
		slog.Debug("Loaded existing agent identity", "agent_id", id, "path", path)
		return id, nil
	}

	// If error is not "not exists", return the error
	if !os.IsNotExist(err) {
		return "", err
	}

	// Generate new UUID
	newID := uuid.New().String()

	// Save to file
	err = os.WriteFile(path, []byte(newID), 0600)
	if err != nil {
		return "", err
	}

	slog.Info("Generated new agent identity", "agent_id", newID, "path", path)
	return newID, nil
}
