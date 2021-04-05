package rest

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/auth"
	"go.uber.org/goleak"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
//TODO: add tests for each method that is calling behind endpoint

func TestDecodeByAlgorithm(t *testing.T) {
	tbl := []struct {
		d   string
		alg string
		res string
	}{
		{"YXNkCg==", "base64", "asd\n" },
		{"", "unknownalgo", "" },
	}
	for i, tt := range tbl {
		actual, err := decodeByAlgorithm([]byte(tt.d), tt.alg)
		if err != nil {
			fmt.Printf("test case with error #%d\n", i)
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, string(actual), tt.res, "test case #%d", i)
	}
}

func TestBase64Decode(t *testing.T) {
	tbl := []struct {
		d   string
		alg string
		res string
	}{
		{"YXNkCg==", "base64", "asd\n" },
		{"bmV3c3RyaW5nCg==", "base64", "newstring\n" },
		{"abracadabra", "base64", "wrongstring" },
	}
	for i, tt := range tbl {
		actual, err := decodeByAlgorithm([]byte(tt.d), tt.alg)
		if err != nil {
			fmt.Printf("test case with error #%d\n", i)
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, string(actual), tt.res, "test case #%d", i)
	}
}

func TestCheckMd5Hash(t *testing.T)  {
	tbl := []struct {
		r   request
	}{
		{request{Encoding: "base64", Data: "MQo=", MD5: "b026324c6904b2a9cb4b88d6d61c81d1"}},
		{request{Encoding: "base64", Data: "MjIK", MD5: "2fc57d6f63a9ee7e2f21a26fa522e3b6"}},
	}

	for i, tt := range tbl {
		err := tt.r.checkMd5Hash()
		assert.Equal(t, err, nil, "test case #%d", i)
	}
}

func TestCheckMd5HashErrors(t *testing.T)  {
	tbl := []struct {
		r   request
		err error
	}{
		{request{Encoding: "base64", Data: "MQo=", MD5: "88d6d61c81d1"}, errors.New("MD5 hash sum is not valid passed: 88d6d61c81d1, calculated: b026324c6904b2a9cb4b88d6d61c81d1")},
		{request{Encoding: "base64", Data: "MQo=1", MD5: "88d6d61c81d1"}, errors.New("illegal base64 data at input byte 4")},
	}
	for i, tt := range tbl {
		assert.Equal(t, tt.r.checkMd5Hash().Error(), tt.err.Error(), "test case #%d", i)
	}
}

func TestCheckJWT(t *testing.T)  {
	_, rest, teardown := startHTTPServer()
	defer teardown()
	tbl := []struct {
		h   string
		err error
	}{
		{"Authorisation Bearer", errors.New("can't parse header: Authorisation contains an invalid number of segments")},
		{"Authorisation", errors.New("can't parse header: Authorisation contains an invalid number of segments")},
		{"", errors.New("can't parse header: Authorisation contains an invalid number of segments")},
		{"Authorisation Bearer 123 123   ", errors.New("can't parse header: Authorisation contains an invalid number of segments")},
	}
	for i, tt := range tbl {
		_, err := rest.checkJWT(tt.h)
		assert.Equal(t, err.Error(), tt.err.Error(), "test case #%d", i)
	}
}


func getRequest(t *testing.T, url string) (data string, statusCode int) {
	resp, err := http.Get(url)
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	return string(body), resp.StatusCode
}

func TestRest_Shutdown(t *testing.T) {
	srv := Rest{}
	done := make(chan bool)

	go func() {
		time.Sleep(200 * time.Millisecond)
		srv.Shutdown()
		close(done)
	}()

	st := time.Now()
	srv.Run(8888)
	assert.True(t, time.Since(st).Seconds() < 1, "should take about 100ms")
	<-done
}

func TestRest_Run(t *testing.T) {
	srv := Rest{}
	port := generateRndPort()
	go func() {
		srv.Run(port)
	}()

	waitHTTPServer(port)

	client := http.Client{}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	srv.Shutdown()
}

func TestRest_Ping(t *testing.T) {
	ts, _, teardown := startHTTPServer()
	defer teardown()

	res, code := getRequest(t, ts.URL+"/ping")
	assert.Equal(t, "pong\n", res)
	assert.Equal(t, 200, code)
}


func startHTTPServer() (ts *httptest.Server, rest *Rest, gracefulTeardown func()) {
	rest = &Rest{
		Version:         "test",
		WorkerServiceURI: "http://localhost:8888/api/v1/",
		Auth:             auth.NewService(auth.Opts{}),
	}
	ts = httptest.NewServer(rest.routes())
	gracefulTeardown = func() {
		ts.Close()
	}
	return ts, rest, gracefulTeardown
}

func generateRndPort() (port int) {
	for i := 0; i < 10; i++ {
		port = 40000 + int(rand.Int31n(10000))
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			_ = ln.Close()
			break
		}
	}
	return port
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

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
