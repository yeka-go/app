package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yeka-go/app"
	"github.com/yeka-go/app/cmd/goapp/internal/cmd/openapi"
)

func init() {
	oapi := &cobra.Command{
		Use:   "openapi",
		Short: "OpenAPI related tools",
	}
	oapi.AddCommand(openapi.MergeCmd, openapi.ServeCmd)
	app.AddCommands(oapi)
}
