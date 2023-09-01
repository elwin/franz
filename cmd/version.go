package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"runtime/debug"
)

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Prints the version of this build",
		RunE: func(cmd *cobra.Command, args []string) error {
			if bi, ok := debug.ReadBuildInfo(); ok {
				for _, setting := range bi.Settings {
					if setting.Key == "vcs.revision" {
						fmt.Println(setting.Value)
						return nil
					}
				}
			}

			return errors.New("Failed to retrieve the version")
		},
	})
}
