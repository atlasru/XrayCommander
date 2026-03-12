package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/atlasru/xraycommander/pkg/models"
	"github.com/pelletier/go-toml/v2"
)

const (
	AppName     = "xraycommander"
	ConfigDir   = ".config/xraycommander"
	ConfigFile  = "config.toml"
	ProfilesDir = "profiles"
)

// Manager handles configuration storage and retrieval
type Manager struct {
	ConfigPath   string
	ProfilesPath string
	DataDir      string
}

// Config represents application configuration
type Config struct {
	XrayPath     string `toml:"xray_path"`
	AutoStart    bool   `toml:"auto_start"`
	LogLevel     string `toml:"log_level"`
	SocksPort    int    `toml:"socks_port"`
	HTTPPort     int    `toml:"http_port"`
	LastProfile  string `toml:"last_profile"`
	SpeedTestURL string `toml:"speed_test_url"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		XrayPath:     "",
		AutoStart:    false,
		LogLevel:     "warning",
		SocksPort:    10808,
		HTTPPort:     10809,
		SpeedTestURL: "https://speed.cloudflare.com/__down?bytes=25000000",
	}
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(home, ConfigDir)
	configPath := filepath.Join(dataDir, ConfigFile)
	profilesPath := filepath.Join(dataDir, ProfilesDir)

	// Create directories if they don't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.MkdirAll(profilesPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	return &Manager{
		ConfigPath:   configPath,
		ProfilesPath: profilesPath,
		DataDir:      dataDir,
	}, nil
}

// LoadConfig loads application configuration
func (m *Manager) LoadConfig() (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			if err := m.SaveConfig(config); err != nil {
				return nil, err
			}
			return config, nil
		}
		return nil, err
	}

	if err := toml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// SaveConfig saves application configuration
func (m *Manager) SaveConfig(config *Config) error {
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	// Write with secure permissions
	file, err := os.OpenFile(m.ConfigPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// GetProfiles returns all saved profiles
func (m *Manager) GetProfiles() ([]models.Profile, error) {
	entries, err := os.ReadDir(m.ProfilesPath)
	if err != nil {
		return nil, err
	}

	var profiles []models.Profile
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.ProfilesPath, entry.Name()))
		if err != nil {
			continue
		}

		var profile models.Profile
		if err := json.Unmarshal(data, &profile); err != nil {
			continue
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// GetProfile loads a specific profile by ID
func (m *Manager) GetProfile(id string) (*models.Profile, error) {
	path := filepath.Join(m.ProfilesPath, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var profile models.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// SaveProfile saves a profile to disk
func (m *Manager) SaveProfile(profile *models.Profile) error {
	if profile.ID == "" {
		profile.ID = uuid.New().String()
	}

	if err := profile.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(m.ProfilesPath, profile.ID+".json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// DeleteProfile removes a profile
func (m *Manager) DeleteProfile(id string) error {
	path := filepath.Join(m.ProfilesPath, id+".json")
	return os.Remove(path)
}

// ExportProfile exports profile as Xray JSON config
func (m *Manager) ExportProfile(id string) ([]byte, error) {
	profile, err := m.GetProfile(id)
	if err != nil {
		return nil, err
	}

	config := profile.ToXrayConfig()
	return json.MarshalIndent(config, "", "  ")
}

// ImportProfile imports a profile from JSON
func (m *Manager) ImportProfile(data []byte) (*models.Profile, error) {
	var profile models.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("invalid profile format: %w", err)
	}

	if profile.ID == "" {
		profile.ID = uuid.New().String()
	}

	if err := m.SaveProfile(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetXrayPath attempts to find Xray binary
func (m *Manager) GetXrayPath() string {
	config, _ := m.LoadConfig()
	if config != nil && config.XrayPath != "" {
		if _, err := os.Stat(config.XrayPath); err == nil {
			return config.XrayPath
		}
	}

	// Search in PATH and common locations
	paths := []string{
		"xray",
		"/usr/local/bin/xray",
		"/usr/bin/xray",
		"/opt/xray/xray",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GetLogPath returns path for log files
func (m *Manager) GetLogPath() string {
	return filepath.Join(m.DataDir, "xray.log")
}

// GetRuntimeConfigPath returns path for temporary Xray config
func (m *Manager) GetRuntimeConfigPath() string {
	return filepath.Join(m.DataDir, "runtime_config.json")
}
