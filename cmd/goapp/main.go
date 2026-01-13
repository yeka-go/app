package main

import (
	"github.com/spf13/cobra"
	"github.com/yeka-go/app"
	_ "github.com/yeka-go/app/internal/cmd"
)

func main() {
	app.SetRootCommand(&cobra.Command{
		Use: "goapp",
	})
	app.Run()
}
