package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sri/learngo/week23/04_cli_app/internal/store"
)

// ========================================
// Root Command — Task Manager CLI
// ========================================
// The root command initializes configuration, sets up the task store,
// and serves as the parent for all subcommands.

var (
	// taskStore is the shared store instance used by all commands.
	taskStore *store.Store

	// cfgFile allows users to specify a custom config file path.
	cfgFile string

	// verbose enables detailed output across all commands.
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "A powerful task manager for your terminal",
	Long: `Tasks is a command-line task manager built with Cobra, Viper, and Bubble Tea.

Manage your tasks from the command line or launch an interactive TUI.
Tasks are stored in a local JSON file for simplicity and portability.

Commands:
  add          Add a new task
  list         List tasks (with optional filters)
  done         Mark a task as completed
  delete       Remove a task
  interactive  Launch the interactive TUI

Examples:
  tasks add "Write documentation" --priority high
  tasks list --status pending
  tasks done 1
  tasks interactive`,

	// PersistentPreRunE runs before ANY subcommand. This is where
	// we initialize the config and the store — ensuring they're
	// available to every command.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeApp()
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("========================================")
		fmt.Println("  Task Manager CLI")
		fmt.Println("========================================")
		fmt.Println()
		fmt.Println("Use 'tasks --help' to see available commands.")
		fmt.Println()
		fmt.Println("Quick start:")
		fmt.Println("  tasks add \"My first task\"")
		fmt.Println("  tasks list")
		fmt.Println("  tasks interactive")
	},
}

// Execute is the entry point called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flags are inherited by all subcommands.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: $HOME/.tasks/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"enable verbose output")

	// Add all subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(interactiveCmd)
}

// initializeApp sets up Viper config and the task store.
func initializeApp() error {
	// ========================================
	// Initialize Viper Configuration
	// ========================================
	if cfgFile != "" {
		// Use the user-specified config file
		viper.SetConfigFile(cfgFile)
	} else {
		// Default: look for config in $HOME/.tasks/ and current directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("finding home directory: %w", err)
		}

		viper.AddConfigPath(filepath.Join(home, ".tasks"))
		viper.AddConfigPath(".")
		viper.AddConfigPath("./04_cli_app")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Set sensible defaults
	home, _ := os.UserHomeDir()
	viper.SetDefault("store.path", filepath.Join(home, ".tasks", "tasks.json"))
	viper.SetDefault("display.show_dates", true)
	viper.SetDefault("display.color", true)
	viper.SetDefault("default_priority", "medium")

	// Read config file (if it exists)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("reading config: %w", err)
		}
		// Config file not found is fine — we use defaults
	} else if verbose {
		fmt.Printf("[verbose] Using config: %s\n", viper.ConfigFileUsed())
	}

	// Environment variable support
	viper.SetEnvPrefix("TASKS")
	viper.AutomaticEnv()

	// ========================================
	// Initialize Task Store
	// ========================================
	storePath := viper.GetString("store.path")
	if verbose {
		fmt.Printf("[verbose] Store path: %s\n", storePath)
	}

	var err error
	taskStore, err = store.New(storePath)
	if err != nil {
		return fmt.Errorf("initializing store: %w", err)
	}

	return nil
}
