package runtime

import (
	"testing"
)

func TestBuiltinSpawnSkipped(t *testing.T) {
	t.Skip("Spawn requires interpreter context and file operations")
}

func TestBuiltinRunSkipped(t *testing.T) {
	t.Skip("Run requires interpreter context and file operations")
}
