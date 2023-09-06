package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Prints the version of this build",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built at: %s\n", date)

			return nil
		},
	})
}
