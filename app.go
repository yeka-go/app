package app

import (
	"context"
	"log"
	"os"
	"os/signal"
)

type ShutdownFunc func(ctx context.Context) error

var shutdownFuncs = make([]ShutdownFunc, 0)

func OnShutdown(funcs ...ShutdownFunc) {
	shutdownFuncs = append(shutdownFuncs, funcs...)
}

func Run() {
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
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
