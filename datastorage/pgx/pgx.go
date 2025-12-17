package pgx

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/yeka-go/app"
)

var instances = make(map[string]*pgx.Conn)

type pgxConfig struct {
	Hosts    string            `mapstructure:"hosts"`
	User     string            `mapstructure:"user"`
	Password string            `mapstructure:"pass"`
	Database string            `mapstructure:"dbname"`
	Options  map[string]string `mapstructure:"options"`
}

func Connect(cmdContext context.Context, connectionName string) (*pgx.Conn, error) {
	conn, ok := instances[connectionName]
	if ok {
		return conn, nil
	}

	configKey := "pgx." + connectionName
	config := app.ConfigFromContext(cmdContext)
	if config == nil || !config.IsSet(configKey) {
		return nil, errors.New("config not found for " + configKey)
	}

	var cfg pgxConfig
	err := config.UnmarshalKey(configKey, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config.UnmarshalKey: %w", err)
	}

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

	conn, err = pgx.ConnectConfig(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	instances[connectionName] = conn
	app.OnShutdown(conn.Close)
	return conn, nil
}
