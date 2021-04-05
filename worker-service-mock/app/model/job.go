package model

import (
	"fmt"
)

type Job struct {
	ID          string `json:"id"`
	TenantID    int    `json:"tenant_id"`
	ClientID    int    `json:"client_id"`
	Payload     string `json:"payload"`
	PayloadSize int    `json:"payload_size"`
	MimeType    string `json:"mime_type"`
}

type JobStatus int

const (
	RUNNING = iota
	SUCCESS
	FAILED
)

func (j JobStatus) String() string {
	switch j {
	case RUNNING:
		return "RUNNING"
	case SUCCESS:
		return "SUCCESS"
	case FAILED:
		return "FAILED"
	default:
		return fmt.Sprintf("%d", int(j))
	}
}
