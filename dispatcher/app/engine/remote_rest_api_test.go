package engine

import (
	"github.com/stretchr/testify/assert"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/model"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/utils"
	"io"
	"testing"
)

func TestRestAPI_GetStatusJob(t *testing.T) {
	repeaterMock := &utils.RepeaterInterfaceMock{
		MakeRequestFunc: func(httpMethod utils.Method, data io.Reader) ([]byte, error) {
			return []byte(`{"status": 1}`), nil
		},
	}
	c := RestAPI{WorkerServiceURL: "http://localhost", Client: repeaterMock,}
	res, err := c.GetStatusJob("1")
	assert.NoError(t, err)
	assert.Equal(t, model.JobStatus(1), res)
	if len(repeaterMock.MakeRequestCalls()) != 1 {
		t.Errorf("[ERROR] makeRequest was called %d times", len(repeaterMock.MakeRequestCalls()))
	}
	t.Logf("%v %T", res, res)
}

func TestRestAPI_SubmitJob(t *testing.T) {
	repeaterMock := &utils.RepeaterInterfaceMock{
		MakeRequestFunc: func(httpMethod utils.Method, data io.Reader) ([]byte, error) {
			return []byte(`{"ID":"4"}`), nil
		},
	}
	c := RestAPI{WorkerServiceURL: "http://localhost", Client: repeaterMock}
	res, err := c.SubmitJob(model.Job{TenantID:1, ClientID:2, Payload:"123", PayloadSize:3})
	assert.NoError(t, err)
	assert.Equal(t, &model.Job{ID:"4"}, res)
	if len(repeaterMock.MakeRequestCalls()) != 1 {
		t.Errorf("[ERROR] makeRequest was called %d times", len(repeaterMock.MakeRequestCalls()))
	}
	t.Logf("%v %T", res, res)
}

func TestRestAPI_GetJob(t *testing.T) {
	repeaterMock := &utils.RepeaterInterfaceMock{
		MakeRequestFunc: func(httpMethod utils.Method, data io.Reader) ([]byte, error) {
			return []byte(`{"status": 2}`), nil
		},
	}
	c := RestAPI{WorkerServiceURL: "http://localhost", Client: repeaterMock}
	res, err := c.GetJob("3")
	assert.NoError(t, err)
	assert.Equal(t, &model.Job{ID: "3", TenantID:3, ClientID:3, Status:"FAILED"}, res)
	if len(repeaterMock.MakeRequestCalls()) != 1 {
		t.Errorf("[ERROR] makeRequest was called %d times", len(repeaterMock.MakeRequestCalls()))
	}
	t.Logf("%v %T", res, res)
}