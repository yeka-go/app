package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yeka-go/app"
	"go.opentelemetry.io/otel"
)

func main() {
	app.SetRootCommand(&cobra.Command{
		Use:   "example",
		Short: "An simple example of application",
		Run: func(cmd *cobra.Command, args []string) {
			_, span := otel.Tracer("").Start(context.Background(), "app start")
			defer span.End()

			fmt.Println("Hello world!")
		},
	})
	app.SetConfigFile("config.yaml")
	app.Run()
}
