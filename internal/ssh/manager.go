package ssh

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

type Manager struct {
	store *store.Store
	emit  func(StatusEvent)

	mu          sync.Mutex
	tunnels     map[string][]*Tunnel
	hosts       []store.SSHHost
	stopHealth  chan struct{}
	healthAlive bool
}

func NewManager(s *store.Store, emit func(StatusEvent)) *Manager {
	return &Manager{
		store:   s,
		emit:    emit,
		tunnels: make(map[string][]*Tunnel),
	}
}

func (m *Manager) Hosts() []store.SSHHost {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]store.SSHHost, len(m.hosts))
	copy(out, m.hosts)
	return out
}

func (m *Manager) LoadHosts(path string) ([]store.SSHHost, error) {
	if path == "" {
		schema, err := m.store.Load()
		if err == nil {
			path = schema.Settings.SSHConfigPath
		}
	}
	hosts, err := ParseConfig(path)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.hosts = hosts
	m.mu.Unlock()
	return hosts, nil
}

func (m *Manager) ConnectAlias(aliasID string) error {
	schema, err := m.store.Load()
	if err != nil {
		return err
	}
	var alias *store.Alias
	for i := range schema.Aliases {
		if schema.Aliases[i].ID == aliasID {
			alias = &schema.Aliases[i]
			break
		}
	}
	if alias == nil {
		return fmt.Errorf("alias %q not found", aliasID)
	}

	if err := m.DisconnectAlias(aliasID); err != nil {
		return err
	}
	if _, err := m.LoadHosts(schema.Settings.SSHConfigPath); err != nil {
		return err
	}

	hosts := m.Hosts()
	tunnels := make([]*Tunnel, 0, len(alias.Forwards))
	for _, fwd := range alias.Forwards {
		var host *store.SSHHost
		for i := range hosts {
			if hosts[i].Name == fwd.SSHHost {
				host = &hosts[i]
				break
			}
		}
		if host == nil {
			return fmt.Errorf("ssh host %q not found in config", fwd.SSHHost)
		}
		t := NewTunnel(aliasID, fwd, *host, m.onTunnelEvent)
		tunnels = append(tunnels, t)
	}

	m.mu.Lock()
	m.tunnels[aliasID] = tunnels
	m.mu.Unlock()

	var wg sync.WaitGroup
	errs := make([]error, len(tunnels))
	for i, t := range tunnels {
		wg.Add(1)
		go func(idx int, tun *Tunnel) {
			defer wg.Done()
			errs[idx] = tun.Connect()
		}(i, t)
	}
	wg.Wait()

	anyFailed := false
	allFailed := true
	for _, e := range errs {
		if e != nil {
			anyFailed = true
		} else {
			allFailed = false
		}
	}

	_, _ = m.store.Update(func(s *store.Schema) {
		for i := range s.Aliases {
			if s.Aliases[i].ID != aliasID {
				continue
			}
			s.Aliases[i].Enabled = !allFailed
			switch {
			case allFailed:
				s.Aliases[i].Status = store.StatusError
			case anyFailed:
				s.Aliases[i].Status = store.StatusError
			default:
				s.Aliases[i].Status = store.StatusConnected
			}
		}
	})

	m.startHealth()

	if allFailed {
		return errors.New("all forwards failed")
	}
	return nil
}

func (m *Manager) DisconnectAlias(aliasID string) error {
	m.mu.Lock()
	tunnels := m.tunnels[aliasID]
	delete(m.tunnels, aliasID)
	empty := len(m.tunnels) == 0
	m.mu.Unlock()

	for _, t := range tunnels {
		_ = t.Disconnect()
	}

	_, _ = m.store.Update(func(s *store.Schema) {
		for i := range s.Aliases {
			if s.Aliases[i].ID == aliasID {
				s.Aliases[i].Status = store.StatusDisconnected
				s.Aliases[i].Enabled = false
			}
		}
	})

	if empty {
		m.stopHealthCheck()
	}
	return nil
}

func (m *Manager) DisconnectAll() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.tunnels))
	for id := range m.tunnels {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		_ = m.DisconnectAlias(id)
	}
}

func (m *Manager) AliasStatus(aliasID string) store.AliasStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.aliasStatusLocked(aliasID)
}

func (m *Manager) aliasStatusLocked(aliasID string) store.AliasStatus {
	tunnels := m.tunnels[aliasID]
	if len(tunnels) == 0 {
		return store.StatusDisconnected
	}
	allConnected := true
	anyConnecting := false
	anyError := false
	for _, t := range tunnels {
		s := t.Status()
		if s != store.StatusConnected {
			allConnected = false
		}
		if s == store.StatusConnecting {
			anyConnecting = true
		}
		if s == store.StatusError {
			anyError = true
		}
	}
	switch {
	case allConnected:
		return store.StatusConnected
	case anyError:
		return store.StatusError
	case anyConnecting:
		return store.StatusConnecting
	default:
		return store.StatusDisconnected
	}
}

func (m *Manager) ConnectedCount(aliasID string) (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	tunnels := m.tunnels[aliasID]
	connected := 0
	for _, t := range tunnels {
		if t.Status() == store.StatusConnected {
			connected++
		}
	}
	return connected, len(tunnels)
}

func (m *Manager) Errors(aliasID string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0)
	for _, t := range m.tunnels[aliasID] {
		if msg := t.LastError(); msg != "" {
			out = append(out, msg)
		}
	}
	return out
}

func (m *Manager) onTunnelEvent(ev StatusEvent) {
	aliasStatus := m.AliasStatus(ev.AliasID)
	_, _ = m.store.Update(func(s *store.Schema) {
		for i := range s.Aliases {
			if s.Aliases[i].ID != ev.AliasID {
				continue
			}
			if s.Aliases[i].Status != aliasStatus {
				s.Aliases[i].Status = aliasStatus
				s.Aliases[i].Enabled = aliasStatus == store.StatusConnected || aliasStatus == store.StatusConnecting
			}
		}
	})
	if m.emit != nil {
		m.emit(ev)
	}
}

func (m *Manager) startHealth() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.healthAlive {
		return
	}
	m.healthAlive = true
	m.stopHealth = make(chan struct{})
	stop := m.stopHealth
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				m.healthTick()
			}
		}
	}()
}

func (m *Manager) stopHealthCheck() {
	m.mu.Lock()
	if !m.healthAlive {
		m.mu.Unlock()
		return
	}
	m.healthAlive = false
	close(m.stopHealth)
	m.mu.Unlock()
}

func (m *Manager) healthTick() {
	m.mu.Lock()
	snapshot := make(map[string][]*Tunnel, len(m.tunnels))
	for k, v := range m.tunnels {
		snapshot[k] = v
	}
	m.mu.Unlock()

	for aliasID, tunnels := range snapshot {
		for _, t := range tunnels {
			if t.Status() == store.StatusConnected && !t.IsAlive() {
				m.onTunnelEvent(StatusEvent{
					AliasID:   aliasID,
					ForwardID: t.ForwardID(),
					Status:    store.StatusError,
					Error:     "process died unexpectedly",
				})
			}
		}
	}
}
