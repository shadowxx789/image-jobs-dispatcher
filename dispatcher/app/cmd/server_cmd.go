package cmd

import (
	"context"
	"github.com/pkg/errors"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/auth"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/engine"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/rest"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ServerCommand struct {
	Version      string
	RemoteEngine EngineGroup `group:"engine" namespace:"engine" env-namespace:"ENGINE"`
	Port         int         `long:"port" env:"SERVER_PORT" default:"9000" description:"Dispatcher server port"`
	CommonOptions
}

type EngineGroup struct {
	Type   string       `long:"type" env:"TYPE" description:"type of storage" choice:"RemoteRest" default:"RemoteRest"`
	Remote RestAPIGroup `group:"Rest" namespace:"Rest" env-namespace:"Rest"`
}

type RestAPIGroup struct {
	WorkerServerURL string `long:"workerServiceAPI" env:"WORKER_SERVICE_API" description:"Remote worker service api url"`
	BlobServerURL   string `long:"blobServiceAPI" env:"BLOB_SERVICE_API" description:"Remote blob service api url"`
}

type application struct {
	*ServerCommand
	rest       *rest.Rest
	terminated chan struct{}
}

//Execute is the entry point for server command
func (sc *ServerCommand) Execute(_ []string) error {
	log.Printf("[INFO] start app server")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] get interrupt signal")
		cancel()
	}()

	app, err := sc.bootstrapApp()
	if err != nil {
		log.Printf("[PANIC] Failed to setup application, %+v", err)
		return err
	}

	if err := app.run(ctx); err != nil {
		log.Printf("[ERROR] server terminated with error %+v", err)
		return err
	}

	log.Printf("[INFO] server terminated")
	return nil
}

func (app *application) run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		app.rest.Shutdown()
		log.Print("[INFO] shutdown is completed")
	}()
	app.rest.Run(app.Port)
	close(app.terminated)
	return nil
}

func (sc *ServerCommand) buildEngine() (result engine.Interface, err error) {
	log.Printf("[INFO] build engine. Type=%s", sc.RemoteEngine.Type)

	switch sc.RemoteEngine.Type {
	case "RemoteRest":
		r := &engine.RestAPI{WorkerServiceURL: sc.WorkerServiceURL}
		return r, nil
	default:
		return nil, errors.Errorf("unsupported engine type %s", sc.RemoteEngine.Type)
	}
	return result, errors.Wrap(err, "can't initialize remote engine")
}

func (sc *ServerCommand) bootstrapApp() (*application, error) {

	engine, err := sc.buildEngine()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build remote engine")
	}

	authService := auth.NewService(auth.Opts{})

	rest := &rest.Rest{
		Version:          sc.Version,
		WorkerServiceURI: sc.WorkerServiceURL,
		RemoteService:    engine,
		Auth:             authService,
	}

	return &application{
		ServerCommand: sc,
		rest:          rest,
		terminated:    make(chan struct{}),
	}, nil
}
