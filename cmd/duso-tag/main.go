package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/duso-org/duso/pkg/version"
)

func main() {
	newTag, err := version.UpdateVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating version: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Tagged:", newTag)

	// Push tag to remote
	cmd := exec.Command("git", "push", "origin", newTag)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to push tag %s: %v\n", newTag, err)
		// Don't exit(1) - tag was created locally, push failure is non-fatal
	} else {
		fmt.Println("Pushed:", newTag)
	}
}
