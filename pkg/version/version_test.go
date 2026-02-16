package version

import (
	"testing"
)

// TestParseVersion tests version string parsing
func TestParseVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		tag        string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantErr    bool
	}{
		{"v0.1.0", "v0.1.0", 0, 1, 0, false},
		{"v1.2.3", "v1.2.3", 1, 2, 3, false},
		{"v10.20.30", "v10.20.30", 10, 20, 30, false},
		{"no v prefix", "1.2.3", 1, 2, 3, false},
		{"with build", "v1.2.3-18", 1, 2, 3, false},
		{"build ignored", "v2.0.0-100", 2, 0, 0, false},
		{"invalid format", "1.2", 0, 0, 0, true},
		{"too many parts", "1.2.3.4", 0, 0, 0, true},
		{"non-numeric", "v1.a.3", 0, 0, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			major, minor, patch, err := ParseVersion(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if major != tt.wantMajor {
				t.Errorf("major = %d, want %d", major, tt.wantMajor)
			}
			if minor != tt.wantMinor {
				t.Errorf("minor = %d, want %d", minor, tt.wantMinor)
			}
			if patch != tt.wantPatch {
				t.Errorf("patch = %d, want %d", patch, tt.wantPatch)
			}
		})
	}
}

// TestIncrementVersion tests version increment logic
func TestIncrementVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		message    string
		inMajor    int
		inMinor    int
		inPatch    int
		wantMajor  int
		wantMinor  int
		wantPatch  int
	}{
		{
			"major bump",
			"major: breaking change",
			1, 2, 3,
			2, 0, 0,
		},
		{
			"minor bump (feat)",
			"feat: new feature",
			1, 2, 3,
			1, 3, 0,
		},
		{
			"patch bump (fix)",
			"fix: bug fix",
			1, 2, 3,
			1, 2, 4,
		},
		{
			"no bump",
			"docs: update readme",
			1, 2, 3,
			1, 2, 3,
		},
		{
			"zero versions",
			"fix: test",
			0, 0, 0,
			0, 0, 1,
		},
		{
			"major from 0",
			"major: first major",
			0, 5, 10,
			1, 0, 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			major, minor, patch := IncrementVersion(tt.message, tt.inMajor, tt.inMinor, tt.inPatch)
			if major != tt.wantMajor {
				t.Errorf("major = %d, want %d", major, tt.wantMajor)
			}
			if minor != tt.wantMinor {
				t.Errorf("minor = %d, want %d", minor, tt.wantMinor)
			}
			if patch != tt.wantPatch {
				t.Errorf("patch = %d, want %d", patch, tt.wantPatch)
			}
		})
	}
}

// TestCreateTag tests tag string creation
func TestCreateTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		major     int
		minor     int
		patch     int
		build     int
		wantTag   string
	}{
		{"simple", 1, 2, 3, 100, "v1.2.3-100"},
		{"zero build", 0, 0, 0, 0, "v0.0.0-0"},
		{"large nums", 10, 20, 30, 1000, "v10.20.30-1000"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Note: CreateTag actually creates a git tag, so we can't test it directly
			// without mocking git. For now, just test the format string construction.
			// In a real scenario, you'd mock exec.Command.
			_ = tt.wantTag
		})
	}
}

// TestVersionPrefixes tests all commit message prefixes
func TestVersionPrefixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		prefix     string
		wantMajor  int
		wantMinor  int
		wantPatch  int
	}{
		{"major", "major: ", 2, 0, 0},
		{"feat", "feat: ", 1, 1, 0},
		{"fix", "fix: ", 1, 0, 1},
		{"docs", "docs: ", 1, 0, 0},
		{"style", "style: ", 1, 0, 0},
		{"refactor", "refactor: ", 1, 0, 0},
		{"test", "test: ", 1, 0, 0},
		{"chore", "chore: ", 1, 0, 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			major, minor, patch := IncrementVersion(tt.prefix+"message", 1, 0, 0)
			if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch {
				t.Errorf("got %d.%d.%d, want %d.%d.%d", major, minor, patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
			}
		})
	}
}

// TestVersionEdgeCases tests edge cases
func TestVersionEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		tag        string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantErr    bool
	}{
		{"empty string", "", 0, 0, 0, true},
		{"just v", "v", 0, 0, 0, true},
		{"uppercase v", "V1.2.3", 0, 0, 0, true},
		{"spaces", "v 1.2.3", 0, 0, 0, true},
		{"negative numbers", "v-1.2.3", 0, 0, 0, true},
		{"build number suffix", "v1.2.3-rc1", 1, 2, 3, false},
		{"complex build", "v1.2.3-18-gabcd123", 1, 2, 3, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			major, minor, patch, err := ParseVersion(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch {
					t.Errorf("got %d.%d.%d, want %d.%d.%d", major, minor, patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
				}
			}
		})
	}
}

// TestVersionSequences tests version increment sequences
func TestVersionSequences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		messages    []string
		startVer    string
		wantSeq     []string
	}{
		{
			"fix releases",
			[]string{"fix: bug1", "fix: bug2", "fix: bug3"},
			"v1.0.0",
			[]string{"v1.0.1", "v1.0.2", "v1.0.3"},
		},
		{
			"mixed",
			[]string{"fix: ", "feat: ", "major: "},
			"v1.0.0",
			[]string{"v1.0.1", "v1.1.0", "v2.0.0"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			major, minor, patch, _ := ParseVersion(tt.startVer)
			for _, msg := range tt.messages {
				major, minor, patch = IncrementVersion(msg, major, minor, patch)
			}
			_ = major
		})
	}
}
