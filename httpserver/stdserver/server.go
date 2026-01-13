package stdserver

import (
	"context"
	"net/http"
	"time"

	"github.com/yeka-go/app/collections"
)

var instances = collections.NewInstances("stdserver", collections.SimpleBuilderWrapper(New), nil)

type Config struct {
	Addr            string        `mapstructure:"address"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	CertFile        string        `mapstructure:"cert_file"`
	CertKeyFile     string        `mapstructure:"cert_key_file"`
}

// Get an already created server instance or create a new server instance from config file based on given serverName
func Get(cmdContext context.Context, serverName string) (*Server, error) {
	return instances.Get(cmdContext, serverName)
}

func New(cfg Config) *Server {

	srv := &Server{
		Server: http.Server{
			Addr: cfg.Addr,
		},
		config: cfg,
	}
	return srv
}

type Server struct {
	Server http.Server
	config Config
}

func (s *Server) Run(appCtx context.Context) error {
	ctx, cancel := context.WithCancel(appCtx)
	defer cancel()

	var err2 error
	go func() {
		if s.config.CertFile != "" {
			err2 = s.Server.ListenAndServeTLS(s.config.CertFile, s.config.CertKeyFile)
		} else {
			err2 = s.Server.ListenAndServe()
			if err2 == http.ErrServerClosed {
				err2 = nil
			}
		}
		if err2 != nil && err2 != http.ErrServerClosed {
			cancel()
		}
	}()

	<-ctx.Done()
	ctxTimeout := context.Background()
	if s.config.ShutdownTimeout > 0 {
		ctt, cancelTimeout := context.WithTimeout(ctxTimeout, s.config.ShutdownTimeout)
		defer cancelTimeout()
		ctxTimeout = ctt
	}
	err := s.Server.Shutdown(ctxTimeout)
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	return nil
}
