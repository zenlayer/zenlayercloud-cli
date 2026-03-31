package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenlayer/zenlayercloud-cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version number of zeno.`,
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Println(version.String())
	return nil
}
