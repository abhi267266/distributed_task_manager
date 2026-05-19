package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLoadOrGenerateID(t *testing.T) {
	tempDir := t.TempDir()
	idFile := filepath.Join(tempDir, ".agent_id")

	// 1. Should generate new ID if file doesn't exist
	id1, err := LoadOrGenerateID(idFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, id1)

	// Verify it's a valid UUID
	_, err = uuid.Parse(id1)
	assert.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(idFile)
	assert.NoError(t, err)
	assert.Equal(t, id1, string(content))

	// 2. Should load existing ID
	id2, err := LoadOrGenerateID(idFile)
	assert.NoError(t, err)
	assert.Equal(t, id1, id2) // Must be exactly the same UUID
}
