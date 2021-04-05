package model

import (
	"fmt"
)

type Job struct {
	ID              string `json:"id"`
	TenantID        int    `json:"tenant_id,omitempty"`
	ClientID        int    `json:"client_id,omitempty"`
	Payload         string `json:"payload,omitempty"`
	PayloadLocation string `json:"payload_location,omitempty"`
	PayloadSize     int    `json:"payload_size,omitempty"`
	Status          string `json:"status,omitempty"`
}

type JobStatus int

const (
	RUNNING = iota
	SUCCESS
	FAILED
	UNKNOWN
)

func (js JobStatus) ToString() string {
	switch js {
	case RUNNING:
		return "RUNNING"
	case SUCCESS:
		return "SUCCESS"
	case FAILED:
		return "FAILED"
	default:
		return fmt.Sprintf("%d", int(js))
	}
}
