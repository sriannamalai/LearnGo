package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ========================================
// Week 23, Lesson 2: Viper Configuration
// ========================================
// Viper is the most popular configuration library in Go. It handles:
//   - Reading from config files (YAML, JSON, TOML, .env, etc.)
//   - Reading from environment variables
//   - Setting default values
//   - Watching config files for live changes
//   - Nested configuration keys
//   - Unmarshaling config into structs
//
// Viper reads configuration in this priority order (highest to lowest):
//   1. Explicit Set calls (viper.Set)
//   2. Flags (when bound to Cobra flags)
//   3. Environment variables
//   4. Config file values
//   5. Key/value store (etcd, Consul)
//   6. Default values
//
// Run this program:
//   go run .
//   APP_PORT=9090 go run .           # Override port via env var
//   APP_DATABASE_HOST=prod-db go run .  # Override nested config via env var
// ========================================

// ========================================
// Config Struct
// ========================================
// Defining a struct for your configuration makes it type-safe
// and enables IDE autocompletion. Viper can unmarshal the config
// file directly into this struct using mapstructure tags.

// AppConfig represents the entire application configuration.
type AppConfig struct {
	App      AppSettings      `mapstructure:"app"`
	Database DatabaseSettings `mapstructure:"database"`
	Logging  LogSettings      `mapstructure:"logging"`
	Features FeatureFlags     `mapstructure:"features"`
}

// AppSettings holds general application settings.
type AppSettings struct {
	Name        string `mapstructure:"name"`
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// DatabaseSettings holds database connection configuration.
type DatabaseSettings struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	Pool     PoolSettings `mapstructure:"pool"`
}

// PoolSettings holds connection pool configuration.
type PoolSettings struct {
	MaxOpen     int           `mapstructure:"max_open"`
	MaxIdle     int           `mapstructure:"max_idle"`
	MaxLifetime time.Duration `mapstructure:"max_lifetime"`
}

// LogSettings holds logging configuration.
type LogSettings struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// FeatureFlags holds feature toggle settings.
type FeatureFlags struct {
	EnableCache      bool `mapstructure:"enable_cache"`
	EnableMetrics    bool `mapstructure:"enable_metrics"`
	EnableRateLimit  bool `mapstructure:"enable_rate_limit"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 23: Viper Configuration")
	fmt.Println("========================================")

	// ========================================
	// 1. Setting Defaults
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("1. Setting Defaults")
	fmt.Println("========================================")

	// Defaults are the lowest priority. They act as fallbacks when
	// no config file, env var, or explicit Set provides a value.
	// Always set sensible defaults for every config key.
	viper.SetDefault("app.name", "MyApp")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", true)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "myapp_dev")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "secret")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.pool.max_open", 10)
	viper.SetDefault("database.pool.max_idle", 5)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetDefault("features.enable_cache", true)
	viper.SetDefault("features.enable_metrics", false)
	viper.SetDefault("features.enable_rate_limit", false)

	fmt.Println("Defaults set for all configuration keys.")
	fmt.Printf("Default app.port: %d\n", viper.GetInt("app.port"))

	// ========================================
	// 2. Reading from Config File
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Reading from Config File")
	fmt.Println("========================================")

	// Tell Viper where to find the config file.
	// Viper supports YAML, JSON, TOML, HCL, .env, and Java properties.
	viper.SetConfigName("config") // Name without extension
	viper.SetConfigType("yaml")   // Explicit file type
	viper.AddConfigPath(".")      // Look in current directory first
	viper.AddConfigPath("./02_viper_config") // Also look in lesson dir

	// ReadInConfig reads the config file and merges it with defaults.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found — using defaults only.")
		} else {
			log.Fatalf("Error reading config file: %v", err)
		}
	} else {
		fmt.Printf("Config file loaded: %s\n", viper.ConfigFileUsed())
	}

	// ========================================
	// 3. Environment Variables
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Environment Variables")
	fmt.Println("========================================")

	// Viper can read from environment variables. This is essential for
	// twelve-factor apps and container deployments where config comes
	// from the environment rather than files.

	// SetEnvPrefix adds a prefix to all env var lookups.
	// With prefix "APP", looking up "port" checks for "APP_PORT".
	viper.SetEnvPrefix("APP")

	// AutomaticEnv makes Viper check env vars for ALL config keys.
	viper.AutomaticEnv()

	// SetEnvKeyReplacer handles nested keys. Viper uses "." as a
	// separator internally (e.g., "database.host"), but env vars
	// use "_" (e.g., "APP_DATABASE_HOST"). The replacer handles
	// this translation.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	fmt.Println("Environment variable support enabled.")
	fmt.Println("Prefix: APP_")
	fmt.Println("Try: APP_PORT=9090 go run .")
	fmt.Println("Try: APP_DATABASE_HOST=prod-db go run .")

	// ========================================
	// 4. Reading Configuration Values
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Reading Configuration Values")
	fmt.Println("========================================")

	// Viper provides typed getters for all common types.
	// These resolve the value using the priority chain:
	//   Set > Flag > Env > Config File > Default
	fmt.Printf("app.name:        %s\n", viper.GetString("app.name"))
	fmt.Printf("app.port:        %d\n", viper.GetInt("app.port"))
	fmt.Printf("app.environment: %s\n", viper.GetString("app.environment"))
	fmt.Printf("app.debug:       %t\n", viper.GetBool("app.debug"))

	// Nested keys use dot notation
	fmt.Printf("\ndatabase.host:   %s\n", viper.GetString("database.host"))
	fmt.Printf("database.port:   %d\n", viper.GetInt("database.port"))
	fmt.Printf("database.name:   %s\n", viper.GetString("database.name"))

	// Deep nesting works too
	fmt.Printf("\ndatabase.pool.max_open: %d\n", viper.GetInt("database.pool.max_open"))
	fmt.Printf("database.pool.max_idle: %d\n", viper.GetInt("database.pool.max_idle"))

	// ========================================
	// 5. Checking If Keys Exist
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Checking If Keys Exist")
	fmt.Println("========================================")

	// IsSet returns true if the key has any value (including from defaults).
	fmt.Printf("app.port is set:     %t\n", viper.IsSet("app.port"))
	fmt.Printf("nonexistent is set:  %t\n", viper.IsSet("nonexistent.key"))

	// AllKeys returns all keys that have been set (from any source).
	allKeys := viper.AllKeys()
	fmt.Printf("Total config keys:   %d\n", len(allKeys))

	// ========================================
	// 6. Explicit Set — Highest Priority
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Explicit Set (Overrides Everything)")
	fmt.Println("========================================")

	// viper.Set has the highest priority. It overrides config files,
	// env vars, and defaults. Use it sparingly — typically for values
	// computed at runtime.
	fmt.Printf("Before Set: app.port = %d\n", viper.GetInt("app.port"))
	viper.Set("app.port", 3000)
	fmt.Printf("After Set:  app.port = %d\n", viper.GetInt("app.port"))

	// Reset for the rest of the demo
	viper.Set("app.port", nil)

	// ========================================
	// 7. Unmarshaling into Structs
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. Unmarshaling into Structs")
	fmt.Println("========================================")

	// Viper can unmarshal the entire config (or a sub-tree) into
	// a struct. This gives you type-safe access and IDE support.
	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	fmt.Printf("Config struct:\n")
	fmt.Printf("  App Name:    %s\n", config.App.Name)
	fmt.Printf("  App Port:    %d\n", config.App.Port)
	fmt.Printf("  Environment: %s\n", config.App.Environment)
	fmt.Printf("  DB Host:     %s\n", config.Database.Host)
	fmt.Printf("  DB Name:     %s\n", config.Database.Name)
	fmt.Printf("  Log Level:   %s\n", config.Logging.Level)
	fmt.Printf("  Cache:       %t\n", config.Features.EnableCache)

	// ========================================
	// 8. Sub-configurations
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. Sub-configurations")
	fmt.Println("========================================")

	// Sub returns a new Viper instance scoped to a sub-tree.
	// Useful when passing config to subsystems that don't need
	// the entire configuration.
	dbConfig := viper.Sub("database")
	if dbConfig != nil {
		fmt.Printf("Database sub-config:\n")
		fmt.Printf("  host: %s\n", dbConfig.GetString("host"))
		fmt.Printf("  port: %d\n", dbConfig.GetInt("port"))
		fmt.Printf("  name: %s\n", dbConfig.GetString("name"))
	}

	// ========================================
	// 9. Watching Config File for Changes
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("9. Watching Config File for Changes")
	fmt.Println("========================================")

	// WatchConfig uses fsnotify to watch the config file for changes.
	// When the file changes, Viper automatically re-reads it.
	// OnConfigChange lets you react to changes (e.g., update log level).
	//
	// This is especially useful for long-running services where you
	// want to change behavior without restarting.

	viper.OnConfigChange(func(e interface{ Name() string }) {
		fmt.Printf("  [watcher] Config file changed: %s\n", e.Name())
		// In a real app, you might:
		//   - Update log level
		//   - Refresh connection pools
		//   - Toggle feature flags
	})

	// Uncomment this line in a long-running application:
	// viper.WatchConfig()

	fmt.Println("Config watching is available for long-running services.")
	fmt.Println("Use viper.WatchConfig() to enable live config reloading.")

	// ========================================
	// 10. Writing Config Files
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("10. Writing Config Files")
	fmt.Println("========================================")

	// Viper can also write configuration files — useful for
	// saving user preferences or generating default configs.
	fmt.Println("Viper can write config files with:")
	fmt.Println("  viper.WriteConfig()         — writes to the file it read from")
	fmt.Println("  viper.SafeWriteConfig()     — writes only if file doesn't exist")
	fmt.Println("  viper.WriteConfigAs(path)   — writes to a specific path")

	// Example: Write current config to a temp file
	tmpFile := os.TempDir() + "/myapp-config-example.yaml"
	if err := viper.WriteConfigAs(tmpFile); err != nil {
		fmt.Printf("  (Could not write example: %v)\n", err)
	} else {
		fmt.Printf("  Example config written to: %s\n", tmpFile)
		// Clean up
		os.Remove(tmpFile)
	}

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println("- viper.SetDefault(): set fallback values")
	fmt.Println("- viper.ReadInConfig(): load from file")
	fmt.Println("- viper.AutomaticEnv(): enable env var reading")
	fmt.Println("- viper.Get*(): type-safe value retrieval")
	fmt.Println("- viper.Unmarshal(): load into structs")
	fmt.Println("- viper.Sub(): scope to config sub-tree")
	fmt.Println("- viper.WatchConfig(): live reload")
	fmt.Println("- Priority: Set > Flags > Env > File > Defaults")
}
