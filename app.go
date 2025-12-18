package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ShutdownFunc func(ctx context.Context) error

var shutdownFuncs = make([]ShutdownFunc, 0)

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
				log.Printf("Waiting for application to stop gracefully, or %v again to terminate the application\n", signal)
				stop()
			case 2:
				log.Println("Terminating application")
				os.Exit(1)
			}
		}
	}()
	return ctx, stop
}

func Run() {
	appCtx, stop := doubleKill()
	defer stop()

	err := executeCommand(appCtx)
	if err != nil {
		log.Println(err)
	}

	shutdownCtx := context.TODO() // TODO should there be a mechanism to set this context, such as adding timeout, etc
	for _, fn := range shutdownFuncs {
		err = fn(shutdownCtx)
		if err != nil {
			log.Println(err)
		}
	}
}
