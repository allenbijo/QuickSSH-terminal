package store

import (
	"os"
	"path/filepath"
	"runtime"
)

const productName = "Quick SSH"

// ConfigFilePath returns the path to the JSON file written by the Electron
// app's electron-store. It matches Electron's app.getPath('userData') +
// "/config.json" on each platform so both apps share the same file.
func ConfigFilePath() string {
	return filepath.Join(userDataDir(), "config.json")
}

func userDataDir() string {
	switch runtime.GOOS {
	case "windows":
		if v := os.Getenv("APPDATA"); v != "" {
			return filepath.Join(v, productName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", productName)
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", productName)
	default:
		if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
			return filepath.Join(v, productName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", productName)
	}
}

func DefaultSSHConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config")
}
