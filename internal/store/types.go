package store

type AliasStatus string

const (
	StatusDisconnected AliasStatus = "disconnected"
	StatusConnecting   AliasStatus = "connecting"
	StatusConnected    AliasStatus = "connected"
	StatusError        AliasStatus = "error"
)

type PortForward struct {
	ID         string `json:"id"`
	LocalPort  int    `json:"localPort"`
	RemoteHost string `json:"remoteHost"`
	RemotePort int    `json:"remotePort"`
	SSHHost    string `json:"sshHost"`
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
