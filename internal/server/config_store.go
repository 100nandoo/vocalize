package server

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// vocalizeConfig is the top-level shape of vocalize.toml.
type vocalizeConfig struct {
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
// api_keys.json (JSON) and write the new vocalize.toml (TOML) without a
// separate legacy struct.
type storedKey struct {
	ID         string     `toml:"id"           json:"id"`
	Name       string     `toml:"name"         json:"name"`
	Prefix     string     `toml:"prefix"       json:"prefix"`
	Hash       string     `toml:"hash"         json:"hash"`
	CreatedAt  time.Time  `toml:"created_at"   json:"createdAt"`
	LastUsedAt *time.Time `toml:"last_used_at" json:"lastUsedAt,omitempty"`
}

// fileMu serialises all reads and writes to vocalize.toml so that the
// summarizer-config and api-key subsystems never overwrite each other's section.
var fileMu sync.Mutex

// vocalizeConfigPath returns the path to vocalize.toml, honouring
// VOCALIZE_CONFIG_DIR when set.
func vocalizeConfigPath() string {
	if dir := os.Getenv("VOCALIZE_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "vocalize.toml")
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "vocalize.toml"
	}
	return filepath.Join(base, "vocalize", "vocalize.toml")
}

// configDir returns just the directory that holds vocalize config files.
func configDir() string {
	if dir := os.Getenv("VOCALIZE_CONFIG_DIR"); dir != "" {
		return dir
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(base, "vocalize")
}

// readVocalizeConfigUnlocked reads vocalize.toml from disk.
// The caller MUST hold fileMu.
func readVocalizeConfigUnlocked() vocalizeConfig {
	var cfg vocalizeConfig
	data, err := os.ReadFile(vocalizeConfigPath())
	if err == nil {
		_ = toml.Unmarshal(data, &cfg)
	}
	return cfg
}

// writeVocalizeConfigUnlocked atomically writes cfg to vocalize.toml.
// The caller MUST hold fileMu.
func writeVocalizeConfigUnlocked(cfg vocalizeConfig) error {
	path := vocalizeConfigPath()
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
