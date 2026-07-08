package store

type AliasStatus string

const (
	StatusDisconnected AliasStatus = "disconnected"
	StatusConnecting   AliasStatus = "connecting"
	StatusConnected    AliasStatus = "connected"
	StatusError        AliasStatus = "error"
)

// ForwardKind distinguishes a classic local (-L) forward from a dynamic
// SOCKS (-D) proxy.
type ForwardKind string

const (
	ForwardLocal   ForwardKind = "local"   // ssh -L localPort:remoteHost:remotePort
	ForwardDynamic ForwardKind = "dynamic" // ssh -D localPort (SOCKS proxy)
)

type PortForward struct {
	ID   string      `json:"id"`
	Kind ForwardKind `json:"kind,omitempty"` // empty == local, for backward compatibility
	// LocalPort is the port opened on the local machine. For a local forward
	// it is the source of the tunnel; for a dynamic forward it is the SOCKS
	// listen port.
	LocalPort int `json:"localPort"`
	// RemoteHost and RemotePort are only used by local forwards.
	RemoteHost string `json:"remoteHost,omitempty"`
	RemotePort int    `json:"remotePort,omitempty"`
	SSHHost    string `json:"sshHost"`
}

// IsDynamic reports whether the forward is a dynamic SOCKS proxy (-D).
func (f PortForward) IsDynamic() bool {
	return f.Kind == ForwardDynamic
}

type Alias struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Forwards []PortForward `json:"forwards"`
	Enabled  bool          `json:"enabled"`
	Status   AliasStatus   `json:"status"`
}

type Settings struct {
	SSHConfigPath         string `json:"sshConfigPath"`
	DefaultLocalPortStart int    `json:"defaultLocalPortStart"`
}

type SSHHost struct {
	Name         string
	HostName     string
	Port         int
	User         string
	IdentityFile string
	ProxyJump    string
}

type Schema struct {
	Settings Settings `json:"settings"`
	Aliases  []Alias  `json:"aliases"`
}

func DefaultSchema() Schema {
	return Schema{
		Settings: Settings{
			SSHConfigPath:         "",
			DefaultLocalPortStart: 9000,
		},
		Aliases: []Alias{},
	}
}
