package cmd

import (
	"strings"
)

type CommonOptionsCommander interface {
	SetCommon(commonOpts CommonOptions)
	Execute(args []string) error
}

type CommonOptions struct {
	WorkerServiceURL string
	BlobServiceURL    string
}

func (c *CommonOptions) SetCommon(commonOpts CommonOptions) {
	c.WorkerServiceURL = strings.TrimSuffix(commonOpts.WorkerServiceURL, "/")
	c.BlobServiceURL = strings.TrimSuffix(commonOpts.BlobServiceURL, "/")
}
