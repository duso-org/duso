//go:build !unix

package runtime

import "os/exec"

// Windows has no process group we can signal the way unix does, so cleanup is
// best effort: the child we started dies, anything it started may outlive it.
// exec() still returns control to the script on time, which is the part that
// has to hold everywhere.

func setProcessGroup(cmd *exec.Cmd) {}

func signalProcessGroup(cmd *exec.Cmd, kill bool) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
