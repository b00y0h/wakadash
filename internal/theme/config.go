package theme

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadThemeFromConfig reads the theme name from ~/.wakatime.cfg.
// Returns an empty string if no theme is set (first run) or if the file doesn't exist.
func LoadThemeFromConfig() (string, error) {
	configPath, err := configFilePath()
	if err != nil {
		return "", err
	}

	// #nosec G304 - configPath is ~/.wakatime.cfg
	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // First run - no config file yet
		}
		return "", fmt.Errorf("cannot open %s: %w", configPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and blank lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "theme" {
			return value, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading config: %w", err)
	}

	return "", nil // No theme key found (first run)
}

// SaveThemeToConfig writes the theme name to ~/.wakatime.cfg.
// Updates an existing theme= line or appends if not found.
func SaveThemeToConfig(themeName string) error {
	configPath, err := configFilePath()
	if err != nil {
		return err
	}

	// Read entire file
	// #nosec G304 - configPath is ~/.wakatime.cfg
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot read %s: %w", configPath, err)
	}

	lines := strings.Split(string(data), "\n")
	updated := false

	// Find and update existing theme= line
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "theme") {
			if key, _, found := strings.Cut(trimmed, "="); found && strings.TrimSpace(key) == "theme" {
				lines[i] = fmt.Sprintf("theme = %s", themeName)
				updated = true
				break
			}
		}
	}

	// Append if not found
	if !updated {
		// Ensure file ends with newline before appending
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("theme = %s", themeName))
	}

	// Write back with user read/write only permissions
	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

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
