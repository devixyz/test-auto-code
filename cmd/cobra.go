package cmd

import (
	"errors"
	"fmt"
	"github.com/Arxtect/Einstein/cmd/api"
	"github.com/Arxtect/Einstein/cmd/config"
	"github.com/Arxtect/Einstein/cmd/version"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:          "Einstein",
	Short:        "Einstein",
	SilenceUsage: true,
	Long:         `Einstein`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			tip()
			return errors.New("parameter error")
		}
		return nil
	},
	PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	Run: func(cmd *cobra.Command, args []string) {
		tip()
	},
}

func tip() {
	usageStr := ` 🚀 Can use` + `-h` + ` View command`
	fmt.Printf("%s\n", usageStr)

}

func init() {
	rootCmd.AddCommand(api.StartCmd)
	rootCmd.AddCommand(version.StartCmd)
	rootCmd.AddCommand(config.StartCmd)

}

// Execute : apply commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
