package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ========================================
// Week 7, Lesson 4 (Mini-Project): Config Reader
// ========================================
// This project builds a configuration reader that:
// 1. Reads a JSON config file
// 2. Parses it into well-defined Go structs
// 3. Provides sensible defaults for missing values
// 4. Validates required fields
// 5. Displays the final configuration
//
// Run: go run main.go
// (Reads config.json from the same directory)
// ========================================

// ========================================
// Configuration Structs
// ========================================

// Config is the top-level configuration structure.
type Config struct {
	AppName        string         `json:"app_name"`
	Version        string         `json:"version"`
	Server         ServerConfig   `json:"server"`
	Database       DatabaseConfig `json:"database"`
	Logging        LoggingConfig  `json:"logging"`
	Features       FeatureFlags   `json:"features"`
	AllowedOrigins []string       `json:"allowed_origins"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`  // seconds
	WriteTimeout int    `json:"write_timeout"` // seconds
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Name           string `json:"name"`
	User           string `json:"user"`
	Password       string `json:"password"`
	MaxConnections int    `json:"max_connections"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level     string `json:"level"`
	File      string `json:"file"`
	MaxSizeMB int    `json:"max_size_mb"`
}

// FeatureFlags holds boolean feature toggles.
type FeatureFlags struct {
	EnableCache         bool `json:"enable_cache"`
	EnableMetrics       bool `json:"enable_metrics"`
	EnableNotifications bool `json:"enable_notifications"`
}

// ========================================
// Validation Errors
// ========================================

// ValidationError represents a single validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors collects multiple validation failures.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no errors"
	}
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = fmt.Sprintf("  - %s", e.Error())
	}
	return fmt.Sprintf("%d validation error(s):\n%s", len(ve), strings.Join(msgs, "\n"))
}

func main() {
	fmt.Println("========================================")
	fmt.Println("Config Reader — Mini-Project")
	fmt.Println("========================================")

	// ========================================
	// Step 1: Read the config file
	// ========================================
	fmt.Println("\n--- Step 1: Reading config file ---")

	configPath := "config.json"
	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("\nFalling back to all defaults...")
		config = DefaultConfig()
	}

	// ========================================
	// Step 2: Apply defaults for missing values
	// ========================================
	fmt.Println("\n--- Step 2: Applying defaults ---")
	ApplyDefaults(&config)
	fmt.Println("  Defaults applied for any missing values")

	// ========================================
	// Step 3: Validate required fields
	// ========================================
	fmt.Println("\n--- Step 3: Validating config ---")
	validationErrors := Validate(config)
	if len(validationErrors) > 0 {
		fmt.Printf("  Validation FAILED:\n%s\n", validationErrors.Error())
		fmt.Println("\n  Please fix the config file and try again.")
		// In a real app, you might os.Exit(1) here.
		// For this demo, we continue to show the config.
	} else {
		fmt.Println("  All validations passed!")
	}

	// ========================================
	// Step 4: Display the final configuration
	// ========================================
	fmt.Println("\n--- Step 4: Final Configuration ---")
	DisplayConfig(config)

	// ========================================
	// Bonus: Demonstrate missing config file handling
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Bonus: Missing Config File")
	fmt.Println("========================================")

	_, err = LoadConfig("nonexistent.json")
	if err != nil {
		fmt.Printf("  Expected error: %v\n", err)
	}

	// ========================================
	// Bonus: Demonstrate partial config
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Bonus: Partial Config (defaults fill gaps)")
	fmt.Println("========================================")

	// Simulate a minimal config file with only a few fields set
	minimalJSON := `{
		"app_name": "MinimalApp",
		"server": {
			"port": 3000
		}
	}`

	var partial Config
	json.Unmarshal([]byte(minimalJSON), &partial)
	ApplyDefaults(&partial)

	fmt.Println("  Minimal config with defaults applied:")
	fmt.Printf("    App:           %s\n", partial.AppName)
	fmt.Printf("    Version:       %s\n", partial.Version)
	fmt.Printf("    Server host:   %s\n", partial.Server.Host)
	fmt.Printf("    Server port:   %d\n", partial.Server.Port)
	fmt.Printf("    DB host:       %s\n", partial.Database.Host)
	fmt.Printf("    DB port:       %d\n", partial.Database.Port)
	fmt.Printf("    Log level:     %s\n", partial.Logging.Level)

	fmt.Println("\n========================================")
	fmt.Println("Done!")
	fmt.Println("========================================")
}

// ========================================
// Core Functions
// ========================================

// LoadConfig reads and parses a JSON config file into a Config struct.
func LoadConfig(path string) (Config, error) {
	var config Config

	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return config, fmt.Errorf("config file not found: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}
	fmt.Printf("  Read %d bytes from %s\n", len(data), path)

	// Parse JSON
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("error parsing JSON: %w", err)
	}
	fmt.Println("  JSON parsed successfully")

	return config, nil
}

// DefaultConfig returns a Config with all default values.
func DefaultConfig() Config {
	return Config{
		AppName: "DefaultApp",
		Version: "0.0.1",
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		Database: DatabaseConfig{
			Host:           "localhost",
			Port:           5432,
			Name:           "app_db",
			User:           "postgres",
			MaxConnections: 10,
		},
		Logging: LoggingConfig{
			Level:     "info",
			File:      "app.log",
			MaxSizeMB: 50,
		},
		Features: FeatureFlags{
			EnableCache:         true,
			EnableMetrics:       false,
			EnableNotifications: false,
		},
		AllowedOrigins: []string{"http://localhost"},
	}
}

// ApplyDefaults fills in zero-valued fields with sensible defaults.
// This ensures the application can run even with a partial config.
func ApplyDefaults(config *Config) {
	defaults := DefaultConfig()

	// App-level defaults
	if config.AppName == "" {
		config.AppName = defaults.AppName
	}
	if config.Version == "" {
		config.Version = defaults.Version
	}

	// Server defaults
	if config.Server.Host == "" {
		config.Server.Host = defaults.Server.Host
	}
	if config.Server.Port == 0 {
		config.Server.Port = defaults.Server.Port
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = defaults.Server.ReadTimeout
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = defaults.Server.WriteTimeout
	}

	// Database defaults
	if config.Database.Host == "" {
		config.Database.Host = defaults.Database.Host
	}
	if config.Database.Port == 0 {
		config.Database.Port = defaults.Database.Port
	}
	if config.Database.Name == "" {
		config.Database.Name = defaults.Database.Name
	}
	if config.Database.User == "" {
		config.Database.User = defaults.Database.User
	}
	if config.Database.MaxConnections == 0 {
		config.Database.MaxConnections = defaults.Database.MaxConnections
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = defaults.Logging.Level
	}
	if config.Logging.File == "" {
		config.Logging.File = defaults.Logging.File
	}
	if config.Logging.MaxSizeMB == 0 {
		config.Logging.MaxSizeMB = defaults.Logging.MaxSizeMB
	}

	// Allowed origins default
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = defaults.AllowedOrigins
	}
}

// Validate checks that required fields are present and values are valid.
// Returns a list of validation errors (empty if valid).
func Validate(config Config) ValidationErrors {
	var errors ValidationErrors

	// Required fields
	if config.AppName == "" {
		errors = append(errors, ValidationError{
			Field: "app_name", Message: "is required"})
	}
	if config.Version == "" {
		errors = append(errors, ValidationError{
			Field: "version", Message: "is required"})
	}

	// Server validation
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		errors = append(errors, ValidationError{
			Field:   "server.port",
			Message: fmt.Sprintf("must be between 1 and 65535, got %d", config.Server.Port),
		})
	}
	if config.Server.ReadTimeout < 1 {
		errors = append(errors, ValidationError{
			Field: "server.read_timeout", Message: "must be at least 1 second"})
	}
	if config.Server.WriteTimeout < 1 {
		errors = append(errors, ValidationError{
			Field: "server.write_timeout", Message: "must be at least 1 second"})
	}

	// Database validation
	if config.Database.Name == "" {
		errors = append(errors, ValidationError{
			Field: "database.name", Message: "is required"})
	}
	if config.Database.User == "" {
		errors = append(errors, ValidationError{
			Field: "database.user", Message: "is required"})
	}
	if config.Database.Port < 1 || config.Database.Port > 65535 {
		errors = append(errors, ValidationError{
			Field:   "database.port",
			Message: fmt.Sprintf("must be between 1 and 65535, got %d", config.Database.Port),
		})
	}
	if config.Database.MaxConnections < 1 {
		errors = append(errors, ValidationError{
			Field: "database.max_connections", Message: "must be at least 1"})
	}

	// Logging validation
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[config.Logging.Level] {
		errors = append(errors, ValidationError{
			Field:   "logging.level",
			Message: fmt.Sprintf("must be debug/info/warn/error, got '%s'", config.Logging.Level),
		})
	}

	return errors
}

// DisplayConfig prints the full configuration in a readable format.
func DisplayConfig(config Config) {
	fmt.Println("\n  Application:")
	fmt.Printf("    Name:           %s\n", config.AppName)
	fmt.Printf("    Version:        %s\n", config.Version)

	fmt.Println("\n  Server:")
	fmt.Printf("    Host:           %s\n", config.Server.Host)
	fmt.Printf("    Port:           %d\n", config.Server.Port)
	fmt.Printf("    Read Timeout:   %ds\n", config.Server.ReadTimeout)
	fmt.Printf("    Write Timeout:  %ds\n", config.Server.WriteTimeout)

	fmt.Println("\n  Database:")
	fmt.Printf("    Host:           %s\n", config.Database.Host)
	fmt.Printf("    Port:           %d\n", config.Database.Port)
	fmt.Printf("    Name:           %s\n", config.Database.Name)
	fmt.Printf("    User:           %s\n", config.Database.User)
	if config.Database.Password != "" {
		fmt.Printf("    Password:       %s\n", strings.Repeat("*", len(config.Database.Password)))
	} else {
		fmt.Printf("    Password:       (not set)\n")
	}
	fmt.Printf("    Max Connections: %d\n", config.Database.MaxConnections)

	fmt.Println("\n  Logging:")
	fmt.Printf("    Level:          %s\n", config.Logging.Level)
	fmt.Printf("    File:           %s\n", config.Logging.File)
	fmt.Printf("    Max Size:       %d MB\n", config.Logging.MaxSizeMB)

	fmt.Println("\n  Features:")
	fmt.Printf("    Cache:          %v\n", config.Features.EnableCache)
	fmt.Printf("    Metrics:        %v\n", config.Features.EnableMetrics)
	fmt.Printf("    Notifications:  %v\n", config.Features.EnableNotifications)

	fmt.Println("\n  Allowed Origins:")
	for _, origin := range config.AllowedOrigins {
		fmt.Printf("    - %s\n", origin)
	}

	// Also show as pretty JSON
	fmt.Println("\n  Full config as JSON:")
	pretty, _ := json.MarshalIndent(config, "  ", "  ")
	fmt.Printf("  %s\n", string(pretty))
}
