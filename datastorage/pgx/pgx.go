package pgx

import (
	"context"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/yeka-go/app"
	"github.com/yeka-go/app/collections"
)

var instances = collections.NewInstances("pgx", New, shutdownPgx)

type Config struct {
	Hosts    string            `mapstructure:"hosts"`
	User     string            `mapstructure:"user"`
	Password string            `mapstructure:"pass"`
	Database string            `mapstructure:"dbname"`
	Options  map[string]string `mapstructure:"options"`
}

// Get an already created pgx connection or create one from config file based on given connectionName
func Get(cmdContext context.Context, connectionName string) (*pgx.Conn, error) {
	return instances.Get(cmdContext, connectionName)
}

func New(cfg Config) (*pgx.Conn, error) {
	q := url.Values{}
	for k, v := range cfg.Options {
		q.Add(k, v)
	}

	dsn := url.URL{
		Scheme:   "postgres",
		Host:     cfg.Hosts,
		User:     url.UserPassword(cfg.User, cfg.Password),
		Path:     cfg.Database,
		RawQuery: q.Encode(),
	}

	conf, _ := pgx.ParseConfig(dsn.String())
	conf.Tracer = &tracer{dbname: conf.Database}

	conn, err := pgx.ConnectConfig(context.Background(), conf)
	return conn, err
}

func shutdownPgx(conn *pgx.Conn) {
	app.OnShutdown(conn.Close)
}
