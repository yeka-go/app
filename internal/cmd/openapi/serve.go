package openapi

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yeka-go/app/internal/openapi/merger"
	"github.com/yeka-go/app/internal/openapi/ui"
	"github.com/yeka-go/app/httpserver/stdserver"
)

var serveTemplate string
var baseUrl string

var ServeCmd = &cobra.Command{
	Use:   "serve <file>",
	Short: "run a webserver for swagger ui",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		res, err := merger.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}

		opt := ui.Option{
			Spec:     res,
			SpecName: filepath.Base(args[0]),
			BaseURL:  baseUrl,
		}

		switch serveTemplate {
		case "", "swagger", "swagger-ui":
			opt.Template = ui.SwaggerUITemplate
		case "redoc", "redocly":
			opt.Template = ui.RedoclyTemplate
		default:
			fmt.Println("unknown template:", serveTemplate)
			return
		}

		srv := stdserver.New(stdserver.Config{Addr: ":8123"})
		srv.Server.Handler = ui.NewHandler(opt)
		fmt.Println("starting server on", srv.Server.Addr)
		if err := srv.Run(cmd.Context()); err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	ServeCmd.Flags().StringVarP(&serveTemplate, "template", "t", "redocly", "template to use: redocly, swagger-ui")
	ServeCmd.Flags().StringVarP(&baseUrl, "baseurl", "b", "", "baseurl to serve the spec (eg: /docs/)")
}
