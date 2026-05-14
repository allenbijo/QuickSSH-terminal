package ssh

import (
	"os"
	"strconv"
	"strings"

	sshcfg "github.com/kevinburke/ssh_config"

	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

func ParseConfig(path string) ([]store.SSHHost, error) {
	if path == "" {
		path = store.DefaultSSHConfigPath()
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []store.SSHHost{}, nil
		}
		return nil, err
	}
	defer f.Close()

	cfg, err := sshcfg.Decode(f)
	if err != nil {
		return nil, err
	}

	hosts := make([]store.SSHHost, 0, len(cfg.Hosts))
	for _, h := range cfg.Hosts {
		for _, pat := range h.Patterns {
			name := pat.String()
			if name == "" || strings.ContainsAny(name, "*?") {
				continue
			}
			port := 22
			if v, _ := cfg.Get(name, "Port"); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					port = n
				}
			}
			hostname, _ := cfg.Get(name, "HostName")
			if hostname == "" {
				hostname = name
			}
			user, _ := cfg.Get(name, "User")
			identity, _ := cfg.Get(name, "IdentityFile")
			proxy, _ := cfg.Get(name, "ProxyJump")
			hosts = append(hosts, store.SSHHost{
				Name:         name,
				HostName:     hostname,
				Port:         port,
				User:         user,
				IdentityFile: identity,
				ProxyJump:    proxy,
			})
		}
	}
	return hosts, nil
}
