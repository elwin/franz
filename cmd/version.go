package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Prints the version of this build",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\n", BuildVersion.Tag)
			fmt.Printf("Commit: %s\n", BuildVersion.Commit)
			fmt.Printf("Built at: %s\n", BuildVersion.Date)

			return nil
		},
	})
}
