package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// ========================================
// Version Command
// ========================================
// A version command is essential for any CLI tool. It helps users
// report issues and verify they're running the expected version.
//
// In production, these values are typically set at build time using
// ldflags:
//   go build -ldflags "-X cmd.Version=1.2.3 -X cmd.BuildDate=2025-01-01"

// These variables can be overridden at build time with ldflags.
var (
	Version   = "0.1.0"
	BuildDate = "2025-01-01"
	GitCommit = "development"
)

// shortVersion controls whether to print only the version number.
var shortVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of MyApp",
	Long: `Display the version, build date, and Git commit of MyApp.

Use --short to print only the version number.

Examples:
  myapp version
  myapp version --short`,

	// Aliases allow "myapp ver" or "myapp v" to work as well.
	Aliases: []string{"ver"},

	// No positional arguments expected.
	Args: cobra.NoArgs,

	Run: func(cmd *cobra.Command, args []string) {
		if shortVersion {
			// Short format: just the version number
			fmt.Println(Version)
			return
		}

		// Full version information
		fmt.Println("========================================")
		fmt.Println("  MyApp Version Information")
		fmt.Println("========================================")
		fmt.Printf("  Version:    %s\n", Version)
		fmt.Printf("  Build Date: %s\n", BuildDate)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)

		if verbose {
			fmt.Println()
			fmt.Println("[verbose] Additional runtime info:")
			fmt.Printf("[verbose]   NumCPU: %d\n", runtime.NumCPU())
			fmt.Printf("[verbose]   GOROOT: %s\n", runtime.GOROOT())
			fmt.Printf("[verbose]   Compiler: %s\n", runtime.Compiler)
		}
	},
}

func init() {
	// Local flag: --short / -s
	versionCmd.Flags().BoolVarP(&shortVersion, "short", "s", false,
		"print only the version number")
}
