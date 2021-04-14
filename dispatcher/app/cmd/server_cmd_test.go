package cmd

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServerApp(t *testing.T) {
	app, ctx, cancel := buildListCmdOpts(t, func(o ServerCommand) ServerCommand {
		return o
	})

	go func() { _ = app.run(ctx) }()
	waitHTTPServer(app.Port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", app.Port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong\n", string(body))

	cancel()
	app.Wait()
}

func createAppFromCmd(t *testing.T, cmd ServerCommand) (*application, context.Context, context.CancelFunc) {
	app, err := cmd.bootstrapApp()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	rand.Seed(time.Now().UnixNano())
	return app, ctx, cancel
}

func waitHTTPServer(port int) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 1)
		conn, _ := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Millisecond*10)
		if conn != nil {
			_ = conn.Close()
			break
		}
	}
}

func buildListCmdOpts(t *testing.T, fn func(o ServerCommand) ServerCommand) (*application, context.Context, context.CancelFunc) {
	cmd := ServerCommand{}
	cmd.SetCommon(CommonOptions{WorkerServiceURL: "http://localhost:8081/api/v1/", BlobServiceURL: "http://localhost:8080/api/v1/"})
	p := flags.NewParser(&cmd, flags.Default)
	_, err := p.ParseArgs([]string{"--port=4356"})
	require.NoError(t, err)
	cmd = fn(cmd)
	return createAppFromCmd(t, cmd)
}

func TestMain(m *testing.M) {
	//Unknown reasons
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("net/http.(*Server).Shutdown"))
}