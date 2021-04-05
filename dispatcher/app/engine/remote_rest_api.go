package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/model"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/utils"
	"log"
	"strconv"
	"sync"
)

type RestAPI struct {
	WorkerServiceURL string
	lock             sync.Mutex
	Client           *utils.Repeater
}

type JobStatusResponse struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Details string `json:"details"`
}

type JobResponse struct {
	model.Job
	Error   string `json:"error"`
	Details string `json:"details"`
}

var store = map[string]model.Job{
	"1": {ID: "1", TenantID: 1, ClientID: 1, PayloadLocation: "/blob/api/v1/1" },
	"2": {ID: "2", TenantID: 2, ClientID: 2},
	"3": {ID: "3", TenantID: 3, ClientID: 3},
}

//SubmitJob submit new image job
func (r *RestAPI) SubmitJob(job model.Job) (*model.Job, error) {
	body, err := json.Marshal(job)
	if err != nil {
		log.Printf("[ERROR] can not encode response body %#v", err)
		return nil, nil
	}
	r.Client = &utils.Repeater{
		ClientTimeout: 10,
		Attempts:      10,
		URI:           r.WorkerServiceURL + "/job",
		Count:         3,
	}
	res, err := r.Client.MakeRequest(utils.POST, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[ERROR] can not make request to get status with error: %#v", err)
		return nil, err
	}
	jsr := &JobResponse{}
	if err = json.NewDecoder(bytes.NewReader(res)).Decode(&jsr); err != nil {
		log.Printf("[ERROR] can not decode response body %#v", err)
		return nil, err
	}

	if jsr.Error != "" {
		err := errors.New(jsr.Error)
		return nil, errors.Wrap(err, jsr.Details)
	}
	r.lock.Lock()	//Workaround only for PoC do not get collision in case of parallel request
	jsr.Job.ID = strconv.Itoa(len(store) + 1)
	r.lock.Unlock()
	jsr.Job.TenantID = job.TenantID
	jsr.Job.ClientID = job.ClientID
	store[jsr.Job.ID] = jsr.Job
	return &model.Job{ID: jsr.ID}, nil
}

//GetJob get job object
func (r *RestAPI) GetJob(id string) (*model.Job, error) {
	status, err := r.GetStatusJob(id)
	//In case if job created by API put status as static RUNNING
	if err != nil {
		log.Printf("[ERROR] can not make request to get status with id: %s, error: %#v", id, err)
		status = 1
	}
	if _, ok := store[id]; !ok {
		log.Printf("[ERROR] no job with id: %s, error: %#v", id, ok)
		return nil, fmt.Errorf("no job with id: %s", id)
	}
	job := store[id]
	if job.Status != status.ToString() {
		job.Status = status.ToString()
	}
	return &job, nil
}

//GetStatusJob get job status
func (r *RestAPI) GetStatusJob(id string) (model.JobStatus, error) {
	r.Client = &utils.Repeater{
		ClientTimeout: 10,
		Attempts:      10,
		URI:           r.WorkerServiceURL + "/job/" + id + "/status",
		Count:         3,
	}
	res, err := r.Client.MakeRequest(utils.GET, nil)
	if err != nil {
		log.Printf("[ERROR] can not make request to get status with id: %s, error: %#v", id, err)
		return -1, err
	}

	jsr := &JobStatusResponse{}
	if err = json.NewDecoder(bytes.NewReader(res)).Decode(&jsr); err != nil {
		log.Printf("[ERROR] can not decode response body %#v", err)
		return -1, err
	}

	if jsr.Error != "" {
		err := errors.New(jsr.Error)
		return -1, errors.Wrap(err, jsr.Details)
	}
	return model.JobStatus(jsr.Status), nil
}
