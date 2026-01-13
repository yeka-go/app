package stdserver

import (
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
)

// Command returns a command to start a server
func Command(cmdName string, serverConfigName string, handler http.Handler) *cobra.Command {
	var listenAddress string
	cmd := &cobra.Command{
		Use: cmdName,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var srv *Server
			if serverConfigName == "" {
				srv = New(Config{Addr: listenAddress})
			} else {
				srv, err = Get(cmd.Context(), serverConfigName)
			}

			if err != nil {
				return err
			}

			if listenAddress != "" {
				srv.Server.Addr = listenAddress
			}
			if srv.Server.Addr == "" {
				srv.Server.Addr = ":8080"
			}

			srv.Server.Handler = handler
			slog.Info("Starting server at " + srv.Server.Addr)
			return srv.Run(cmd.Context())
		},
	}
	cmd.Flags().StringVarP(&listenAddress, "address", "a", "", "Listen address, will override setting from config file (default: :8080)")
	return cmd
}
