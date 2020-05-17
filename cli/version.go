package cli

import (
	"fmt"
	"os"

	"github.com/clintjedwards/toolkit/version"
	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev_000000_333333"

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print version number",
	Run:   runVersionCmd,
}

func runVersionCmd(cmd *cobra.Command, args []string) {
	info, err := version.Parse(appVersion)
	if err != nil {
		fmt.Printf("could not parse version: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Markdown Server v%s %s %s\n", info.Semver, info.Epoch, info.Hash)
}

func init() {
	RootCmd.AddCommand(cmdVersion)
}
