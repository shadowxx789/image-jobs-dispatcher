package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJobStatus_ToString(t *testing.T) {
	tbl := []struct {
		js   JobStatus
		res string
	}{
		{JobStatus(0), "RUNNING"},
		{JobStatus(1), "SUCCESS"},
		{JobStatus(2), "FAILED"},
		{JobStatus(15), "15"},

	}
	for i, tt := range tbl {
		assert.Equal(t, tt.res, tt.js.ToString(), "test case #%d", i)
	}
}
