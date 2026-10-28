package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"printer/internal/book"
)

// Config is remembered between runs.
type Config struct {
	Settings     book.Settings `json:"settings"`
	LastFile     string        `json:"lastFile"`
	Printer      string        `json:"printer"`
	ReverseBacks bool          `json:"reverseBacks"` // manual duplex: feed backs last-to-first
	RotateBacks  bool          `json:"rotateBacks"`  // manual duplex: turn backs 180°
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "book-printer", "config.json"), nil
}

func loadConfig() Config {
	cfg := Config{Settings: book.DefaultSettings()}
	path, err := configPath()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	if json.Unmarshal(data, &cfg) != nil || cfg.Settings.Validate() != nil {
		return Config{Settings: book.DefaultSettings()}
	}
	return cfg
}

func saveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
