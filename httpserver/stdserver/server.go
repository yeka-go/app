package stdserver

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
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
	Server      http.Server
	config      Config
	connCounter atomic.Int64 // TODO for metric (http.server.open_connections)
}

// Run the server
// If appCtx is cancelled, graceful shutdown will be triggered
// and will wait for ShutdownTimeout duration before forcefully kill the server.
func (s *Server) Run(appCtx context.Context) error {
	ctx, cancel := context.WithCancel(appCtx)
	defer cancel()

	s.Server.ConnState = func(c net.Conn, cs http.ConnState) {
		switch cs {
		case http.StateNew:
			s.connCounter.Add(1)
		case http.StateClosed:
			s.connCounter.Add(-1)
		}
	}

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
