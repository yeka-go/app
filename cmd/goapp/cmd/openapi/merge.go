package openapi

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yeka-go/app/cmd/goapp/internal/openapi/merger"
)

var MergeCmd = &cobra.Command{
	Use:   "merge <file>",
	Short: "merge openapi documents",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		res, err := merger.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(res))
	},
}
