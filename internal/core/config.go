package core

import (
	"os"
	"path/filepath"
)

// configDirName is the per-user directory (under $HOME) where Conduit keeps its
// vault, connection profiles, history DB, and keystore.
const configDirName = ".conduit"

// ConfigDir returns the absolute path to the per-user config directory,
// creating it (0700) if necessary.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, configDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// ConfigPath joins name onto the config directory, ensuring the directory
// exists. Use it for well-known files (vault.enc, connections.json, history.db).
func ConfigPath(name string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}
