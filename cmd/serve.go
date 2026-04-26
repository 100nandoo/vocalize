package cmd

import (
	"embed"
	"fmt"

	"github.com/100nandoo/inti/internal/server"
	"github.com/spf13/cobra"
)

// WebFS is set by main package via embed.go
var WebFS embed.FS

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server with web UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		if port != 0 {
			cfg.Port = port
		}
		host, _ := cmd.Flags().GetString("host")
		if host != "" {
			cfg.Host = host
		}
		fmt.Printf("starting server at http://%s:%d\n", cfg.Host, cfg.Port)
		return server.Start(cfg, WebFS)
	},
}

func init() {
	serveCmd.Flags().Int("port", 0, "HTTP port (default 8282)")
	serveCmd.Flags().String("host", "", "HTTP host (default 127.0.0.1)")
}
