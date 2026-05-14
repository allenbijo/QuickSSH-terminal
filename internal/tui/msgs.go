package tui

import (
	"github.com/allenbijo/QuickSSH-terminal/internal/ssh"
	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

type statusMsg ssh.StatusEvent

type schemaLoadedMsg struct {
	schema store.Schema
}

type hostsLoadedMsg struct {
	hosts []store.SSHHost
	err   error
}

type connectResultMsg struct {
	aliasID string
	err     error
}

type disconnectDoneMsg struct {
	aliasID string
}

type sessionEndedMsg struct {
	err error
}

type errorMsg struct {
	err error
}

type clearFlashMsg struct{}

type screen int

const (
	screenList screen = iota
	screenEditor
	screenSettings
	screenConfirm
	screenHostPicker
)
