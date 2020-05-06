package cli

import (
	"github.com/clintjedwards/mdserver/mdserver"
	"github.com/spf13/cobra"
)

// cmdServer manipulates topics
var cmdServer = &cobra.Command{
	Use:   "server <url>",
	Short: "Launches a webserver that displays markdown files",
	Long:  "<url> is the <host>:<port> that webserver will listen on",
	Args:  cobra.MaximumNArgs(1),
	Run:   runServerCmd,
}

func runServerCmd(cmd *cobra.Command, args []string) {
	url := args[0]
	dir, _ := cmd.Flags().GetString("directory")

	mdserver.Run(dir, url, "")
}

func init() {
	cmdServer.Flags().StringP("directory", "d", ".", "directory of markdown files")
	cmdServer.Flags().StringP("open", "o", "", "name of the file to open to, if not provided a default home page will be provided")

	RootCmd.AddCommand(cmdServer)
}
