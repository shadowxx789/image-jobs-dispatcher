package main

import (
	"context"
	"github.com/theshamuel/image-jobs-dispatcher/worker-service-mock/app/rest"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var version = "unknown"

type application struct {
	rest        *rest.Rest
	terminated  chan struct{}
}

func (app *application) run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		app.rest.Shutdown()
		log.Print("[INFO] shutdown is completed")
	}()

	app.rest.Run()
	close(app.terminated)
}

func main()  {
	rest := &rest.Rest{}

	app := &application{
		rest:          rest,
		terminated:    make(chan struct{}),
	}
	log.Printf("[INFO] starting Worker Service API mock server version %s", version)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] get interrupt signal")
		cancel()
	}()

	app.run(ctx)
	log.Printf("[INFO] mock server terminated")
}
