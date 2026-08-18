//go:build unix

package runtime

import (
	"os/exec"
	"syscall"
)

// setProcessGroup puts the child in its own process group, so a timeout can
// signal the whole tree. Signalling just the child leaks its grandchildren,
// which is exactly what happens with the wrapper scripts people actually run.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// signalProcessGroup sends SIGTERM, or SIGKILL when kill is set, to the group
// created by setProcessGroup. The negative pid is what fans it out to the group.
func signalProcessGroup(cmd *exec.Cmd, kill bool) {
	if cmd.Process == nil {
		return
	}
	sig := syscall.SIGTERM
	if kill {
		sig = syscall.SIGKILL
	}
	_ = syscall.Kill(-cmd.Process.Pid, sig)
}
