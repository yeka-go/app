package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yeka-go/app"
	"github.com/yeka-go/app/datastorage/pgx"
	"go.opentelemetry.io/otel"
)

func main() {
	app.SetRootCommand(&cobra.Command{
		Use:   "pgx",
		Short: "An example of using pgx to connect to postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := otel.Tracer("").Start(context.Background(), "app start")
			defer span.End()

			db, err := pgx.Connect(cmd.Context(), "example")
			if err != nil {
				return fmt.Errorf("pgx.Connect: %w", err)
			}

			err = db.Ping(ctx)
			if err != nil {
				return fmt.Errorf("db.Ping: %w", err)
			}

			_, err = db.Exec(ctx, "CREATE TABLE IF NOT EXISTS users (id int, name varchar)")
			if err != nil {
				return fmt.Errorf("db.Exec: %w", err)
			}

			res, err := db.Query(ctx, "SELECT * FROM users")
			if err != nil {
				return fmt.Errorf("db.Query: %w", err)
			}
			res.Close()

			return nil
		},
	})

	app.SetConfigFile("config.yaml")
	app.Run()
}
