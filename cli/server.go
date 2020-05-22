package cli

import (
	"github.com/clintjedwards/mdserver/mdserver"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// cmdServer manipulates topics
var cmdServer = &cobra.Command{
	Use:   "server <url>",
	Short: "Launches a webserver that displays markdown files",
	Long:  "<url> is the <host>:<port> that webserver will listen on",
	Args:  cobra.MinimumNArgs(1),
	Run:   runServerCmd,
}

func runServerCmd(cmd *cobra.Command, args []string) {
	url := args[0]
	dir, _ := cmd.Flags().GetString("directory")
	theme, _ := cmd.Flags().GetString("theme")

	server := mdserver.NewMDServer(mdserver.ServerOptions{
		Dir:   dir,
		Addr:  url,
		Open:  "",
		Theme: theme,
	})

	log.Fatal().Err(server.Run()).Msg("server exited")
}

func init() {
	cmdServer.Flags().StringP("directory", "d", ".", "directory of markdown files")
	cmdServer.Flags().StringP("open", "o", "", "name of the file to open to, if not provided a default home page will be provided")
	cmdServer.Flags().StringP("theme", "t", "dark", "css theme; supports values 'dark' or 'light'")

	RootCmd.AddCommand(cmdServer)
}
