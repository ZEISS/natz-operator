package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var config = DefaultConfig()

const (
	versionFmt = "%s (%s %s)"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	RootCmd.AddCommand(CredsCmd)

	RootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", config.Verbose, "verbose output")
	RootCmd.PersistentFlags().BoolVarP(&config.Force, "force", "f", config.Force, "force overwrite")

	CredsCmd.PersistentFlags().StringVarP(&config.Creds.User, "user", "u", config.Creds.User, "user name")

	RootCmd.SilenceErrors = true
	RootCmd.SilenceUsage = true
}

var RootCmd = &cobra.Command{
	Use:   "nctl",
	Short: "nctl is a tool for managing operator resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRoot(cmd.Context())
	},
	Version: fmt.Sprintf(versionFmt, version, commit, date),
}

func runRoot(_ context.Context) error {
	return nil // no-op
}
