package server

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// intiConfig is the top-level shape of inti.toml.
type intiConfig struct {
	Summarizer summarizerSection `toml:"summarizer"`
	APIKeys    []storedKey       `toml:"api_keys"`
}

// summarizerSection holds the persisted summarizer configuration.
type summarizerSection struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	APIKey   string `toml:"api_key"`
}

// storedKey is a single hashed API key entry.
// Both toml and json tags are kept so the migration code can unmarshal old
// api_keys.json (JSON) and write the new inti.toml (TOML) without a
// separate legacy struct.
type storedKey struct {
	ID         string     `toml:"id"           json:"id"`
	Name       string     `toml:"name"         json:"name"`
	Prefix     string     `toml:"prefix"       json:"prefix"`
	Hash       string     `toml:"hash"         json:"hash"`
	CreatedAt  time.Time  `toml:"created_at"   json:"createdAt"`
	LastUsedAt *time.Time `toml:"last_used_at" json:"lastUsedAt,omitempty"`
}

// fileMu serialises all reads and writes to inti.toml so that the
// summarizer-config and api-key subsystems never overwrite each other's section.
var fileMu sync.Mutex

// intiConfigPath returns the path to inti.toml, honouring
// INTI_CONFIG_DIR when set.
func intiConfigPath() string {
	if dir := os.Getenv("INTI_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "inti.toml")
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "inti.toml"
	}
	return filepath.Join(base, "inti", "inti.toml")
}

// configDir returns just the directory that holds inti config files.
func configDir() string {
	if dir := os.Getenv("INTI_CONFIG_DIR"); dir != "" {
		return dir
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(base, "inti")
}

// readIntiConfigUnlocked reads inti.toml from disk.
// The caller MUST hold fileMu.
func readIntiConfigUnlocked() intiConfig {
	var cfg intiConfig
	data, err := os.ReadFile(intiConfigPath())
	if err == nil {
		_ = toml.Unmarshal(data, &cfg)
	}
	return cfg
}

// writeIntiConfigUnlocked atomically writes cfg to inti.toml.
// The caller MUST hold fileMu.
func writeIntiConfigUnlocked(cfg intiConfig) error {
	path := intiConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	enc := toml.NewEncoder(f)
	encErr := enc.Encode(cfg)
	f.Close()
	if encErr != nil {
		os.Remove(tmp)
		return encErr
	}
	return os.Rename(tmp, path)
}
