// Copyright (c) 2022 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dockertests

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
)

// Opens a docker pool connection and create a btcd network.
func createBtcdNetwork(t *testing.T) (*dockertest.Pool, *dockertest.Network) {
	// Connect to docker (a pool is a connection to docker)
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	network, err := pool.CreateNetwork(
		"btcd_dockertests_network",
		func(config *dc.CreateNetworkOptions) {
			config.Internal = true
		})
	if err != nil {
		t.Fatalf("Could not created network: %s", err)
	}

	return pool, network
}

// Purges the provided network.
func purgeNetwork(t *testing.T, network *dockertest.Network) {
	if err := network.Close(); err != nil {
		t.Fatalf("Could not purge resource: %s", err)
	}
}

// Starts the btcd container with default options (no RPC server).
func startBtcdDefaultOpts(t *testing.T, pool *dockertest.Pool, network *dockertest.Network) *dockertest.Resource {
	options := &dockertest.RunOptions{
		Repository: "btcd-dockertests",
		Tag:        "latest",
		Networks:   []*dockertest.Network{network},
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	return resource
}

// Starts the btcd container with RPC server enabled.
func startBtcdWithRPC(t *testing.T, pool *dockertest.Pool, network *dockertest.Network) *dockertest.Resource {
	options := &dockertest.RunOptions{
		Repository: "btcd-dockertests",
		Tag:        "latest",
		Networks:   []*dockertest.Network{network},
		Cmd: []string{
			"--rpcuser=localuser",
			"--rpcpass=localuserpwd",
			"--rpclisten=0.0.0.0",
		},
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	return resource
}

// Purges the provided resource/container.
func purgeContainer(t *testing.T, resource *dockertest.Resource) {
	if err := resource.Close(); err != nil {
		t.Fatalf("Could not purge resource: %s", err)
	}
}

// Connects to RPC Server and returns RPC client.
//
// Uses pool.Retry() to first establish a TLS connection to wait for RPC server start.
// RPC Server certificate is obtained and used to estalish RPC client connection.
func rpcConnect(t *testing.T, pool *dockertest.Pool, network *dockertest.Network, resource *dockertest.Resource) *rpcclient.Client {
	rpcServerIp := resource.GetIPInNetwork(network)
	rpcAddress := rpcServerIp + ":8334"

	// Retry TLS Connect since container might be starting up
	var tlsConn *tls.Conn
	var tlsConnErr error
	tlsConnErr = pool.Retry(func() error {
		tlsConn, tlsConnErr = tls.Dial("tcp", rpcAddress, &tls.Config{
			InsecureSkipVerify: true, // Do not verify TLS Certificate since it's self signed.
		})
		if tlsConnErr != nil {
			t.Logf("Failed to TLS connect: " + tlsConnErr.Error())
		}
		return tlsConnErr
	})
	if tlsConnErr != nil {
		t.Fatalf("Failed to TLS connect, err %s", tlsConnErr)
	}

	// Store RPC Server certificates so that rpcclient can use it
	var rpcServerCerts bytes.Buffer
	for _, cert := range tlsConn.ConnectionState().PeerCertificates {
		err := pem.Encode(&rpcServerCerts, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})
		if err != nil {
			t.Fatalf("Failed to save to memory RPC Server certificates, err %s", err)
		}
	}
	tlsConn.Close()

	connCfg := &rpcclient.ConnConfig{
		Host:                rpcAddress,
		Endpoint:            "ws",
		User:                "localuser",
		Pass:                "localuserpwd",
		Certificates:        rpcServerCerts.Bytes(),
		DisableConnectOnNew: true,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		t.Fatalf("Failed to initialize rpc client, err %s", err)
	}

	err = client.Connect(1)
	if err != nil {
		t.Fatalf("Failed to connect to RPC Server, err %s", err)
	}
	return client
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
