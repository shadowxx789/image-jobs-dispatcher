package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/auth"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/engine"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/model"
	"go.uber.org/goleak"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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
		r inputMessage
	}{
		{inputMessage{Encoding: "base64", Data: "MQo=", MD5: "b026324c6904b2a9cb4b88d6d61c81d1"}},
		{inputMessage{Encoding: "base64", Data: "MjIK", MD5: "2fc57d6f63a9ee7e2f21a26fa522e3b6"}},
	}

	for i, tt := range tbl {
		err := tt.r.checkMd5Hash()
		assert.Equal(t, err, nil, "test case #%d", i)
	}
}

func TestCheckMd5HashErrors(t *testing.T)  {
	tbl := []struct {
		r   inputMessage
		err error
	}{
		{inputMessage{Encoding: "base64", Data: "MQo=", MD5: "88d6d61c81d1"}, errors.New("MD5 hash sum is not valid passed: 88d6d61c81d1, calculated: b026324c6904b2a9cb4b88d6d61c81d1")},
		{inputMessage{Encoding: "base64", Data: "MQo=1", MD5: "88d6d61c81d1"}, errors.New("illegal base64 data at input byte 4")},
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
		{"Bearer", errors.New("can't parse header: Authorisation contains an invalid number of segments")},
		{"", errors.New("can't parse header: Authorisation contains an invalid number of segments")},
		{"Authorisation Bearer 123 123   ", errors.New("can't parse header: Authorisation contains an invalid number of segments")},
	}
	for i, tt := range tbl {
		_, err := rest.checkJWT(tt.h)
		assert.Equal(t, err.Error(), tt.err.Error(), "test case #%d", i)
	}
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
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	srv.Shutdown()
}

func TestRest_Ping(t *testing.T) {
	ts, _, teardown := startHTTPServer()
	defer teardown()

	res, code := getRequest(t, ts.URL+"/ping")
	assert.Equal(t, "pong\n", res)
	assert.Equal(t, http.StatusOK, code)
}

func TestRest_GetJobStatus(t *testing.T) {
	ts, r, teardown := startHTTPServer()
	defer teardown()
	r.RemoteService = &engine.InterfaceMock{
		GetStatusJobFunc: func(id string) (model.JobStatus, error) {
			return model.JobStatus(1), nil
		},
	}
	res, code := getRequest(t, ts.URL+"/api/v1/job/1/status")
	assert.Equal(t, "{\"status\":\"SUCCESS\"}", strings.ReplaceAll(res, " ",""))
	assert.Equal(t, http.StatusOK, code)
}

func TestRest_SubmitJob(t *testing.T) {
	ts, r, teardown := startHTTPServer()
	defer teardown()
	r.RemoteService = &engine.InterfaceMock{
		SubmitJobFunc: func(job model.Job) (*model.Job, error) {
			return &model.Job{ID: "4"}, nil
		},
	}
	reqBody, err := json.Marshal(inputMessage{Encoding: "base64", Data: "MQo=", MD5: "b026324c6904b2a9cb4b88d6d61c81d1"})
	if err != nil {
		t.Logf("[ERROR] error to marshal object inside test")
	}
	res, code := postRequest(t, ts.URL+"/api/v1/job", bytes.NewReader(reqBody))
	assert.Equal(t, "{\"id\":\"4\"}", strings.ReplaceAll(res, " ",""))
	assert.Equal(t, http.StatusCreated, code)
}

func TestRest_GetJob(t *testing.T) {
	ts, r, teardown := startHTTPServer()
	defer teardown()
	r.RemoteService = &engine.InterfaceMock{
		GetJobFunc: func(id string) (*model.Job, error) {
			return &model.Job{ID: "3", TenantID:2, ClientID:1, PayloadLocation: "img/1"}, nil
		},
		GetStatusJobFunc: func(id string) (model.JobStatus, error) {
			return model.JobStatus(1), nil
		},
	}
	res, code := getRequest(t, ts.URL+"/api/v1/job/3")
	assert.Equal(t, `{"id":"3","tenant_id":2,"client_id":1,"payload_location":"img/1"}`, strings.ReplaceAll(res, " ",""))
	assert.Equal(t, http.StatusOK, code)
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

func getRequest(t *testing.T, url string) (data string, statusCode int) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJ0aWQiOjEsIm9pZCI6MSwiYXVkIjoiY29tLmNvbXBhbnkuam9ic2VydmljZSIsImF6cCI6IjEiLCJlbWFpbCI6ImN1c3RvbWVyQG1haWwuY29tIn0.CcTapGbWX0UEMovUwC8kAcWMUxmbOeO0qhsu-wqHQH0")
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	return string(body), resp.StatusCode
}

func postRequest(t *testing.T, url string, reqBody io.Reader) (data string, statusCode int) {
	req, err := http.NewRequest("POST", url, reqBody)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJ0aWQiOjEsIm9pZCI6MSwiYXVkIjoiY29tLmNvbXBhbnkuam9ic2VydmljZSIsImF6cCI6IjEiLCJlbWFpbCI6ImN1c3RvbWVyQG1haWwuY29tIn0.CcTapGbWX0UEMovUwC8kAcWMUxmbOeO0qhsu-wqHQH0")
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	return string(body), resp.StatusCode
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
