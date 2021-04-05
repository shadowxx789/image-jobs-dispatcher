package engine

import (
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/model"
)

type Interface interface {
	SubmitJob(job model.Job) (*model.Job, error)
	GetJob(id string) (*model.Job, error)
	GetStatusJob(id string) (model.JobStatus, error)
}
