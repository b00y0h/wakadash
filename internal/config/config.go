// Package config handles reading the WakaTime configuration file (~/.wakatime.cfg).
//
// Attribution: Rewritten from github.com/sahaj-b/wakafetch (MIT License)
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the WakaTime API credentials and endpoint URL.
type Config struct {
	APIURL string
	APIKey string
}

// Load reads ~/.wakatime.cfg and returns the API credentials.
// It normalizes the Wakapi self-hosted URL to the WakaTime-compatible path.
func Load() (*Config, error) {
	configPath, err := configFilePath()
	if err != nil {
		return nil, fmt.Errorf("cannot determine config path: %w", err)
	}

	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", configPath, err)
	}
	defer f.Close()

	cfg := &Config{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and blank lines.
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "api_url":
			cfg.APIURL = value
		case "api_key":
			cfg.APIKey = value
		}

		if cfg.APIURL != "" && cfg.APIKey != "" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	if cfg.APIURL == "" {
		return nil, fmt.Errorf("api_url not found in %s", configPath)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key not found in %s", configPath)
	}

	cfg.APIURL = strings.TrimSuffix(cfg.APIURL, "/")

	// Normalize Wakapi self-hosted URL to WakaTime-compatible endpoint.
	if cfg.APIURL == "https://wakapi.dev/api" {
		cfg.APIURL = "https://wakapi.dev/api/compat/wakatime"
	}

	return cfg, nil
}

// configFilePath returns the absolute path to ~/.wakatime.cfg.
func configFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".wakatime.cfg"), nil
}
