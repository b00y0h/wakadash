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
	APIURL      string
	APIKey      string
	HistoryRepo string // e.g., "b00y0h/wakatime-data"
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
	currentSection := "" // Track current INI section (empty = root/global)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and blank lines.
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section header [section_name]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Parse keys based on current section
		if currentSection == "" {
			// Root/global section: parse api_url and api_key
			switch key {
			case "api_url":
				cfg.APIURL = value
			case "api_key":
				cfg.APIKey = value
			}
		} else if currentSection == "wakadash" {
			// [wakadash] section: parse history_repo
			if key == "history_repo" {
				cfg.HistoryRepo = value
			}
		}

		// Early exit once required fields are found
		// (but continue scanning to find optional HistoryRepo if present)
		if cfg.APIURL != "" && cfg.APIKey != "" && cfg.HistoryRepo != "" {
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

// EnsureWakadashSection adds a [wakadash] section to ~/.wakatime.cfg if missing.
// Called on startup to provide user with documented config options.
func EnsureWakadashSection() error {
	configPath, err := configFilePath()
	if err != nil {
		return err
	}

	// Read entire file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Config doesn't exist yet, will be created by WakaTime plugin
		}
		return fmt.Errorf("cannot read %s: %w", configPath, err)
	}

	content := string(data)

	// Check if [wakadash] section exists (case-insensitive)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section := strings.ToLower(strings.Trim(trimmed, "[]"))
			if section == "wakadash" {
				return nil // Section already exists
			}
		}
	}

	// Append [wakadash] section template
	template := `
[wakadash]
# Theme: dracula, nord, gruvbox, monokai, solarized, tokyonight
# theme = dracula

# GitHub repo for historical data (format: owner/repo)
# Set up your archive with wakasync: https://github.com/b00y0h/wakasync
# history_repo = your-username/wakatime-data
`

	// Ensure file ends with newline before appending
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += template

	// Write back with user read/write only permissions
	return os.WriteFile(configPath, []byte(content), 0600)
}

// configFilePath returns the absolute path to ~/.wakatime.cfg.
func configFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".wakatime.cfg"), nil
}
