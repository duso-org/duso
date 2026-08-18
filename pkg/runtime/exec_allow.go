package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The exec() allowlist. exec() is off unless the operator passes -allow-exec,
// and what that flag names is the whole of what exec() can run.
//
//	-allow-exec "certbot,systemctl start arland,systemctl restart arland"
//
// Entries are comma-separated. Each entry is a command, optionally followed by
// args that a call must supply verbatim as its leading arguments. Bare "certbot"
// permits certbot with any args; "systemctl restart arland" permits only that
// exact invocation (plus anything after it).
//
// Every entry is resolved to an absolute path once, here at startup, and that
// path is what gets executed. There is no PATH lookup at call time, so nothing
// a script or a request can do later changes which binary runs.
//
// The allowlist is not a sandbox and the docs say so. Most real commands can be
// talked into running something else (certbot --pre-hook, git -c core.pager,
// tar --to-command), so trusting a command means trusting everyone who can
// influence its arguments. What the allowlist does buy is that argv can never
// choose the *program* - a compromised handler cannot reach for a shell.
type execRule struct {
	name   string   // command as written in the flag, used for matching and messages
	path   string   // absolute path resolved at startup
	prefix []string // args a call must start with, empty means any args
}

var (
	execEnabled  bool       // -allow-exec was passed in some form
	execAllowAll bool       // -allow-exec with no list: anything goes
	execRules    []execRule // resolved entries, empty when execAllowAll
)

// shellNames are flagged at boot because allowing one is allowing everything.
var shellNames = map[string]bool{
	"sh": true, "bash": true, "zsh": true, "dash": true,
	"ksh": true, "fish": true, "csh": true, "tcsh": true,
}

// InitExecAllowlist reads -allow-exec from the sys datastore and resolves it.
// Called once at startup, before any script runs. A returned error is fatal:
// an operator who asked for a command that isn't there wants to hear about it
// now, not at 3am when the handler first fires.
func InitExecAllowlist() error {
	sysDs := GetDatastore("duso_sys", nil)
	raw, _ := sysDs.Get("-allow-exec")
	if raw == nil {
		return nil
	}
	execEnabled = true

	// A valueless flag parses as boolean true. Nobody writes that by accident,
	// so take it at face value - they're testing something - and be loud.
	list, isString := raw.(string)
	if !isString || strings.TrimSpace(list) == "" {
		execAllowAll = true
		fmt.Fprintln(os.Stderr, "warning: -allow-exec with no command list - exec() can run anything this user can run")
		return nil
	}

	for _, entry := range strings.Split(list, ",") {
		tokens := strings.Fields(entry)
		if len(tokens) == 0 {
			continue // trailing or doubled comma
		}

		path, err := exec.LookPath(tokens[0])
		if err != nil {
			return fmt.Errorf("allow-exec: %s not found", tokens[0])
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("allow-exec: %s: %w", tokens[0], err)
		}

		if shellNames[filepath.Base(path)] {
			fmt.Fprintf(os.Stderr, "warning: -allow-exec permits %s - a shell can run anything\n", tokens[0])
		}

		execRules = append(execRules, execRule{
			name:   tokens[0],
			path:   path,
			prefix: tokens[1:],
		})
	}

	// "-allow-exec ,,," is the valueless case wearing a hat.
	if len(execRules) == 0 {
		execAllowAll = true
		fmt.Fprintln(os.Stderr, "warning: -allow-exec with no command list - exec() can run anything this user can run")
	}
	return nil
}

// resolveExec checks a call against the allowlist and returns the absolute path
// to run. The error is what the script sees, so it names the flag that would
// permit the call.
func resolveExec(command string, argv []string) (string, error) {
	if !execEnabled {
		return "", fmt.Errorf("exec: %s not permitted; start duso with -allow-exec %q", command, command)
	}

	if execAllowAll {
		path, err := exec.LookPath(command)
		if err != nil {
			return "", fmt.Errorf("exec: %s not found", command)
		}
		return path, nil
	}

	// Collect prefix mismatches so a call that names a permitted command but
	// the wrong args gets told what the permitted args are.
	var nameMatched []execRule
	for _, rule := range execRules {
		if rule.name != command && rule.path != command {
			continue
		}
		nameMatched = append(nameMatched, rule)
		if hasPrefix(argv, rule.prefix) {
			return rule.path, nil
		}
	}

	if len(nameMatched) > 0 {
		allowed := make([]string, 0, len(nameMatched))
		for _, rule := range nameMatched {
			allowed = append(allowed, strings.Join(append([]string{rule.name}, rule.prefix...), " "))
		}
		return "", fmt.Errorf("exec: %s not permitted with those args; permitted: %s",
			command, strings.Join(allowed, ", "))
	}

	return "", fmt.Errorf("exec: %s not permitted; start duso with -allow-exec %q", command, command)
}

// hasPrefix reports whether argv begins with every token in prefix.
func hasPrefix(argv, prefix []string) bool {
	if len(argv) < len(prefix) {
		return false
	}
	for i, want := range prefix {
		if argv[i] != want {
			return false
		}
	}
	return true
}
