package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

type StatusEvent struct {
	AliasID   string
	ForwardID string
	Status    store.AliasStatus
	Error     string
}

type Tunnel struct {
	aliasID string
	forward store.PortForward
	host    store.SSHHost
	emit    func(StatusEvent)

	mu       sync.Mutex
	cmd      *exec.Cmd
	status   store.AliasStatus
	stderr   bytes.Buffer
	lastErr  string
	stopping bool
}

func NewTunnel(aliasID string, forward store.PortForward, host store.SSHHost, emit func(StatusEvent)) *Tunnel {
	return &Tunnel{
		aliasID: aliasID,
		forward: forward,
		host:    host,
		emit:    emit,
		status:  store.StatusDisconnected,
	}
}

func (t *Tunnel) Status() store.AliasStatus {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.status
}

func (t *Tunnel) LastError() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastErr
}

func (t *Tunnel) ForwardID() string { return t.forward.ID }

func (t *Tunnel) Connect() error {
	t.mu.Lock()
	if t.status == store.StatusConnected || t.status == store.StatusConnecting {
		t.mu.Unlock()
		return nil
	}
	t.stderr.Reset()
	t.stopping = false
	t.setStatusLocked(store.StatusConnecting, "")

	args := []string{
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=15",
		"-o", "ServerAliveCountMax=3",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=no",
		"-L", fmt.Sprintf("%d:%s:%d", t.forward.LocalPort, t.forward.RemoteHost, t.forward.RemotePort),
	}
	if t.host.Port != 0 && t.host.Port != 22 {
		args = append(args, "-p", strconv.Itoa(t.host.Port))
	}
	if t.host.User != "" {
		args = append(args, "-l", t.host.User)
	}
	if t.host.IdentityFile != "" {
		args = append(args, "-i", expandHome(t.host.IdentityFile))
	}
	if t.host.ProxyJump != "" {
		args = append(args, "-J", t.host.ProxyJump)
	}
	args = append(args, t.host.HostName)

	cmd := exec.Command("ssh", args...)
	cmd.Stderr = &t.stderr
	cmd.Env = append(os.Environ(), "SSH_ASKPASS_REQUIRE=prefer")
	hideWindow(cmd)

	if err := cmd.Start(); err != nil {
		t.setStatusLocked(store.StatusError, err.Error())
		t.mu.Unlock()
		return err
	}
	t.cmd = cmd
	t.mu.Unlock()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		t.mu.Lock()
		defer t.mu.Unlock()
		t.cmd = nil
		if t.stopping {
			t.setStatusLocked(store.StatusDisconnected, "")
			return nil
		}
		msg := strings.TrimSpace(t.stderr.String())
		if msg == "" {
			if err != nil {
				msg = err.Error()
			} else {
				msg = "ssh exited"
			}
		}
		t.setStatusLocked(store.StatusError, msg)
		return errors.New(msg)
	case <-time.After(2500 * time.Millisecond):
		t.mu.Lock()
		defer t.mu.Unlock()
		if t.status == store.StatusConnecting {
			t.setStatusLocked(store.StatusConnected, "")
		}
		go t.watchExit(done)
		return nil
	}
}

func (t *Tunnel) watchExit(done <-chan error) {
	err := <-done
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cmd = nil
	if t.stopping || t.status == store.StatusDisconnected {
		t.setStatusLocked(store.StatusDisconnected, "")
		return
	}
	msg := strings.TrimSpace(t.stderr.String())
	if msg == "" {
		if err != nil {
			msg = err.Error()
		} else {
			msg = "ssh exited"
		}
	}
	t.setStatusLocked(store.StatusError, msg)
}

func (t *Tunnel) Disconnect() error {
	t.mu.Lock()
	cmd := t.cmd
	t.stopping = true
	t.setStatusLocked(store.StatusDisconnected, "")
	t.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	killGroup(cmd)

	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		<-done
	}
	return nil
}

func (t *Tunnel) IsAlive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cmd == nil || t.cmd.Process == nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return t.cmd.Process.Signal(syscall.Signal(0)) == nil
}

func (t *Tunnel) setStatusLocked(status store.AliasStatus, errMsg string) {
	if t.status == status && t.lastErr == errMsg {
		return
	}
	t.status = status
	t.lastErr = errMsg
	if t.emit != nil {
		ev := StatusEvent{
			AliasID:   t.aliasID,
			ForwardID: t.forward.ID,
			Status:    status,
			Error:     errMsg,
		}
		go t.emit(ev)
	}
}

func expandHome(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	return home + p[1:]
}
