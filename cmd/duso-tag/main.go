package main

import (
	"fmt"
	"os"

	"github.com/duso-org/duso/pkg/version"
)

func main() {
	newTag, err := version.UpdateVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating version: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(newTag)
}
