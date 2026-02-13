package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type appConfig struct {
	shutdownTimeout time.Duration
}

var appDefaultConfig = appConfig{}

type RunOption func(*appConfig)

func WithShutdownTimeout(d time.Duration) RunOption {
	return func(ac *appConfig) {
		ac.shutdownTimeout = d
	}
}

type ShutdownFunc func(ctx context.Context) error

var shutdownFuncs = make([]ShutdownFunc, 0)

// OnShutdown registers shutdown functions, which will be called before application exit.
// User should avoid using log.Fatal() or os.Exit() as there's no way to catch it.
// Instead, please use app.Exit()
func OnShutdown(funcs ...ShutdownFunc) {
	shutdownFuncs = append(shutdownFuncs, funcs...)
}

func doubleKill() (context.Context, context.CancelFunc) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		counter := 0
		for {
			sig := <-ch
			counter++
			switch counter {
			case 1:
				signal := "kill"
				if sig == os.Interrupt {
					signal = "press Ctrl+C"
				}
				slog.Warn(fmt.Sprintf("Waiting for application to stop gracefully, or %v again to terminate the application\n", signal))
				stop()
			case 2:
				slog.Warn("Terminating application")
				os.Exit(1)
			}
		}
	}()
	return ctx, stop
}

// Exit calls shutdown functions before calling os.Exit()
func Exit(exitCode int, msg string) {
	shutdown()
	slog.Warn(msg)
	os.Exit(exitCode)
}

func RunWithOptions(opts ...RunOption) {
	for _, fn := range opts {
		fn(&appDefaultConfig)
	}
	Run()
}

func Run() {
	appCtx, stop := doubleKill()
	defer stop()

	err := executeCommand(appCtx)
	if err != nil {
		slog.Error(err.Error())
	}

	shutdown()
}

func shutdown() {
	shutdownCtx := context.Background()
	if appDefaultConfig.shutdownTimeout > 0 {
		ctx, stop := context.WithTimeout(shutdownCtx, appDefaultConfig.shutdownTimeout)
		defer stop()
		shutdownCtx = ctx
	}

	wg := sync.WaitGroup{}
	for _, fn := range shutdownFuncs {
		wg.Go(func() {
			err := fn(shutdownCtx)
			if err != nil {
				slog.Error(err.Error())
			}
		})
	}
	wg.Wait()
}
