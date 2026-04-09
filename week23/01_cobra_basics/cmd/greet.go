package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// ========================================
// Greet Command
// ========================================
// Demonstrates local flags, flag validation, and command aliases.
// Local flags are only available to the command they're defined on,
// unlike persistent flags which propagate to all subcommands.

// Local flag variables for the greet command.
var (
	greetName  string
	greetShout bool
	greetTimes int
)

var greetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Greet someone by name",
	Long: `The greet command prints a personalized greeting.

You can customize the greeting with flags:
  --name   The name to greet (required)
  --shout  Print the greeting in uppercase
  --times  Repeat the greeting N times

Examples:
  myapp greet --name Sri
  myapp greet --name World --shout
  myapp greet --name Go --times 3`,

	// ========================================
	// Aliases
	// ========================================
	// Aliases let users invoke the same command with shorter names.
	// "myapp hi" is the same as "myapp greet".
	Aliases: []string{"hi", "hello"},

	// ========================================
	// Args Validation
	// ========================================
	// Cobra provides built-in validators for positional arguments:
	//   - cobra.NoArgs: command takes no positional arguments
	//   - cobra.ExactArgs(n): command takes exactly n arguments
	//   - cobra.MinimumNArgs(n): at least n arguments
	//   - cobra.MaximumNArgs(n): at most n arguments
	//   - cobra.RangeArgs(min, max): between min and max arguments
	//
	// We use NoArgs here because all input comes through flags.
	Args: cobra.NoArgs,

	// ========================================
	// PreRunE — Validation Before Execution
	// ========================================
	// PreRunE runs before Run and can return an error to abort.
	// This is the ideal place for flag validation.
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if greetName == "" {
			return fmt.Errorf("--name flag is required (use --name <your-name>)")
		}
		if greetTimes < 1 {
			return fmt.Errorf("--times must be at least 1, got %d", greetTimes)
		}
		if greetTimes > 100 {
			return fmt.Errorf("--times must be at most 100, got %d", greetTimes)
		}
		return nil
	},

	// Run is the main function for this command.
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Println("[verbose] Greet command executing")
			fmt.Printf("[verbose] Name: %s, Shout: %t, Times: %d\n",
				greetName, greetShout, greetTimes)
		}

		greeting := fmt.Sprintf("Hello, %s! Welcome to the Cobra CLI.", greetName)

		if greetShout {
			greeting = strings.ToUpper(greeting)
		}

		for i := 0; i < greetTimes; i++ {
			if greetTimes > 1 {
				fmt.Printf("[%d/%d] %s\n", i+1, greetTimes, greeting)
			} else {
				fmt.Println(greeting)
			}
		}
	},
}

func init() {
	// ========================================
	// Local Flags
	// ========================================
	// Local flags are only available on the command they're defined on.
	// They do NOT propagate to subcommands.
	//
	// StringVarP binds the flag to a variable:
	//   StringVarP(&variable, "long-name", "short", default, description)
	//   - "long-name" → --name
	//   - "short"     → -n
	//   - default     → "" (empty string)

	greetCmd.Flags().StringVarP(&greetName, "name", "n", "",
		"name of the person to greet (required)")

	greetCmd.Flags().BoolVarP(&greetShout, "shout", "s", false,
		"print greeting in uppercase")

	greetCmd.Flags().IntVarP(&greetTimes, "times", "t", 1,
		"number of times to repeat the greeting")

	// ========================================
	// Marking Flags as Required
	// ========================================
	// Cobra can enforce that a flag is provided. If the user omits it,
	// Cobra will print an error before the command runs.
	// Note: We also validate in PreRunE for a better error message.
	greetCmd.MarkFlagRequired("name")
}
