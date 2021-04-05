package engine

import (
	"github.com/stretchr/testify/assert"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/model"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/utils"
	"io"
	"testing"
)

func TestRestAPI_GetStatusJob(t *testing.T) {
	c := RestAPI{WorkerServiceURL: "http://localhost", Client: &utils.RepeaterInterfaceMock{
		MakeRequestFunc: func(httpMethod utils.Method, data io.Reader) ([]byte, error) {
			return []byte(`{"status": 1}`), nil
		},
	}}
	res, err := c.GetStatusJob("1")
	assert.NoError(t, err)
	assert.Equal(t, model.JobStatus(1), res)
	t.Logf("%v %T", res, res)
}

func TestRestAPI_SubmitJob(t *testing.T) {
	c := RestAPI{WorkerServiceURL: "http://localhost", Client: &utils.RepeaterInterfaceMock{
		MakeRequestFunc: func(httpMethod utils.Method, data io.Reader) ([]byte, error) {
			return []byte(`{"ID":"4"}`), nil
		},
	}}
	res, err := c.SubmitJob(model.Job{TenantID:1, ClientID:2, Payload:"123", PayloadSize:3})
	assert.NoError(t, err)
	assert.Equal(t, &model.Job{ID:"4"}, res)
	t.Logf("%v %T", res, res)
}

func TestRestAPI_GetJob(t *testing.T) {
	c := RestAPI{WorkerServiceURL: "http://localhost", Client: &utils.RepeaterInterfaceMock{
		MakeRequestFunc: func(httpMethod utils.Method, data io.Reader) ([]byte, error) {
			return []byte(`{"status": 2}`), nil
		},
	}}
	res, err := c.GetJob("3")
	assert.NoError(t, err)
	assert.Equal(t, &model.Job{ID: "3", TenantID:3, ClientID:3, Status:"FAILED"}, res)
	t.Logf("%v %T", res, res)
}