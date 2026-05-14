package ssh

import "os/exec"

// SessionCommand builds the interactive ssh command for the given host name.
// The TUI hands this to tea.ExecProcess so stdio is inherited and the user
// gets a real terminal session.
func SessionCommand(host string) *exec.Cmd {
	return exec.Command("ssh", "-t", "-t", host)
}
