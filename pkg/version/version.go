package version

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GetLastTag returns the most recent git tag
func GetLastTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no git tags found")
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitMessage returns the current commit message
func GetCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitCount returns the total number of commits
func GetCommitCount() (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	return count, err
}

// ParseVersion extracts major.minor.patch from a version tag (e.g., "v0.5.0-18" -> "0.5.0")
func ParseVersion(tag string) (major, minor, patch int, err error) {
	// Remove "v" prefix and anything after dash
	version := strings.TrimPrefix(tag, "v")
	version = strings.Split(version, "-")[0]

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, err
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, err
	}
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, err
	}

	return major, minor, patch, nil
}

// IncrementVersion bumps version based on commit message
func IncrementVersion(message string, major, minor, patch int) (int, int, int) {
	if strings.HasPrefix(message, "major:") {
		return major + 1, 0, 0
	}
	if strings.HasPrefix(message, "feat:") {
		return major, minor + 1, 0
	}
	if strings.HasPrefix(message, "fix:") {
		return major, minor, patch + 1
	}
	// No recognized prefix, no version bump
	return major, minor, patch
}

// CreateTag creates a new git tag with the given version and build number
func CreateTag(major, minor, patch, build int) (string, error) {
	tag := fmt.Sprintf("v%d.%d.%d-%d", major, minor, patch, build)
	cmd := exec.Command("git", "tag", tag)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create tag: %w", err)
	}
	return tag, nil
}

// UpdateVersion handles the full version bump process
func UpdateVersion() (string, error) {
	// Get last tag
	lastTag, err := GetLastTag()
	if err != nil {
		// No tags yet, start at v0.1.0
		lastTag = "v0.1.0"
	}

	// Parse current version
	major, minor, patch, err := ParseVersion(lastTag)
	if err != nil {
		return "", err
	}

	// Get commit message
	message, err := GetCommitMessage()
	if err != nil {
		return "", err
	}

	// Store old version
	oldMajor, oldMinor, oldPatch := major, minor, patch

	// Increment based on message
	major, minor, patch = IncrementVersion(message, major, minor, patch)

	// Only create tag if version changed
	if major == oldMajor && minor == oldMinor && patch == oldPatch {
		return "", nil
	}

	// Get commit count for build
	build, err := GetCommitCount()
	if err != nil {
		return "", err
	}

	// Create the tag
	newTag, err := CreateTag(major, minor, patch, build)
	if err != nil {
		return "", err
	}

	return newTag, nil
}
