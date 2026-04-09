package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ========================================
// Root Command
// ========================================
// Every Cobra application has a root command. This is the command
// that runs when no subcommand is specified. It also serves as the
// parent for all subcommands and is where persistent flags (flags
// inherited by all subcommands) are defined.

// verbose is a persistent flag — available to ALL subcommands.
// Persistent flags propagate down the command tree.
var verbose bool

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	// Use is the one-line usage message.
	// The first word is the command name.
	Use: "myapp",

	// Short is shown in the 'help' output next to the command name.
	Short: "MyApp is a demo CLI built with Cobra",

	// Long is shown when the user runs 'myapp --help'.
	// Use it to provide detailed documentation.
	Long: `MyApp is a demonstration CLI application built with Cobra.

It showcases common Cobra patterns including:
  - Subcommands (greet, version)
  - Persistent flags (--verbose, available to all commands)
  - Local flags (--name, --shout, available only to specific commands)
  - Argument validation
  - Command aliases
  - Auto-generated help text

This is Week 23, Lesson 1 of the LearnGo curriculum.`,

	// Run is the function executed when the root command is called directly
	// (i.e., the user runs "myapp" with no subcommand).
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("========================================")
		fmt.Println("  Welcome to MyApp!")
		fmt.Println("========================================")
		fmt.Println()
		fmt.Println("This is a demo CLI built with Cobra.")
		fmt.Println("Run 'myapp --help' to see available commands.")
		fmt.Println()

		if verbose {
			fmt.Println("[verbose] Verbose mode is ON")
			fmt.Println("[verbose] Root command executed")
			fmt.Println("[verbose] No subcommand specified")
		}
	},
}

// Execute is called by main.main(). It is the entry point for the CLI.
// This function adds all child commands to the root command and sets flags.
// It only needs to happen once in rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints the error, but we ensure a non-zero exit code.
		os.Exit(1)
	}
}

func init() {
	// ========================================
	// Persistent Flags
	// ========================================
	// Persistent flags are available to the command they're defined on
	// AND every subcommand underneath it. Since we define --verbose on
	// the root command, ALL subcommands inherit it.
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"enable verbose output for detailed logging")

	// ========================================
	// Adding Subcommands
	// ========================================
	// Each subcommand is defined in its own file and added here.
	// This keeps the code organized — one file per command.
	rootCmd.AddCommand(greetCmd)
	rootCmd.AddCommand(versionCmd)
}
