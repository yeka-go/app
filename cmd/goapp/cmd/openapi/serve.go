package openapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yeka-go/app/cmd/goapp/internal/openapi/merger"
	"github.com/yeka-go/app/cmd/goapp/internal/openapi/ui"
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

		srv := http.Server{Addr: ":8123", Handler: ui.NewHandler(opt)}
		go func() {
			fmt.Println("starting server")
			err := srv.ListenAndServe()
			if err != nil {
				log.Println(err)
			}
		}()
		<-cmd.Context().Done()
		srv.Shutdown(context.TODO())
	},
}

func init() {
	ServeCmd.Flags().StringVarP(&serveTemplate, "template", "t", "redocly", "template to use: redocly, swagger-ui")
	ServeCmd.Flags().StringVarP(&baseUrl, "baseurl", "b", "", "baseurl to serve the spec (eg: /docs/)")
}
