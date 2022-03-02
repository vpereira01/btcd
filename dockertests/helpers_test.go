// Copyright (c) 2022 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dockertests

import (
	"strings"
	"testing"

	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
)

// Opens a pool and startd a btcd container.
func startBtcd(t *testing.T) (*dockertest.Pool, *dockertest.Resource) {
	// Connect to docker (a pool is a connection to docker)
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository: "btcd-dockertests",
		Tag:        "latest",
		NetworkID:  "none",
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	return pool, resource
}

// Purges the provided resource using pool.
func purgeContainer(t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource) {
	// When you're done, kill and remove the container
	if err := pool.Purge(resource); err != nil {
		t.Fatalf("Could not purge resource: %s", err)
	}
}

// Setups logging for a container so that all logs sent to testing.T.Log().
// For continous logging this should be run as a go routine.
func logContainer(t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource) {
	var stdOutLB = loggerBridge{t, "cont-out> "}
	var stdErrLB = loggerBridge{t, "cont-err> "}

	opts := dc.LogsOptions{
		Stderr:      true,
		Stdout:      true,
		Follow:      true,
		Timestamps:  false,
		RawTerminal: false, // If true joins container stderr and stdout into OutputStream

		Container: resource.Container.ID,

		OutputStream: stdOutLB,
		ErrorStream:  stdErrLB,
	}

	// Get container logs
	pool.Client.Logs(opts)
}

// An io.Writer that redirects writes to testing.T.Log()
type loggerBridge struct {
	t      *testing.T
	prefix string
}

func (lb loggerBridge) Write(p []byte) (n int, err error) {
	// This is a safe conversion since strings are read-only slice of bytes
	// Reference: https://stackoverflow.com/a/34863211
	var s = string(p)

	// Do not log line breaks to avoid noise
	ls := strings.ReplaceAll(s, "\n", "")
	ls = strings.ReplaceAll(ls, "\r", "")

	lb.t.Log(lb.prefix + ls)

	// We must return that we have written all s, otherwise it stops working properly.
	return len(s), nil
}
