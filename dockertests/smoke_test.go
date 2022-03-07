// Copyright (c) 2022 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dockertests

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"net"
	"testing"

	"github.com/btcsuite/btcd/rpcclient"
)

func TestDefaultListeningPorts(t *testing.T) {
	pool, network := createBtcdNetwork(t)
	defer purgeNetwork(t, network)

	resource := startBtcdDefaultOpts(t, pool, network)
	defer purgeContainer(t, resource)

	go logContainer(t, pool, resource)

	btcdIp := resource.GetIPInNetwork(network)

	// Do not fail imediatly when connecting to P2P port since container is still starting
	err := pool.Retry(func() error {
		address := btcdIp + ":8333"
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Logf("Not able to connect to %s, err %s", address, err)
		} else {
			conn.Close()
		}
		return err
	})
	if err != nil {
		t.Fatalf("P2P port not open, err %s", err)
	}

	conn, err := net.Dial("tcp", btcdIp+":8334")
	if err == nil {
		t.Fatalf("RPC port open")
		conn.Close()
	}
}

func TestRPCListeningPort(t *testing.T) {
	pool, network := createBtcdNetwork(t)
	defer purgeNetwork(t, network)

	resource := startBtcdWithRPC(t, pool, network)
	defer purgeContainer(t, resource)

	go logContainer(t, pool, resource)

	btcdIp := resource.GetIPInNetwork(network)

	// Do not fail imediatly when connecting to RPC port since container is still starting
	err := pool.Retry(func() error {
		address := btcdIp + ":8334"
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Logf("Not able to connect to %s, err %s", address, err)
		} else {
			conn.Close()
		}
		return err
	})
	if err != nil {
		t.Fatalf("RPC port not open, err %s", err)
	}
}

func TestRPCUsesTLS(t *testing.T) {
	pool, network := createBtcdNetwork(t)
	defer purgeNetwork(t, network)

	resource := startBtcdWithRPC(t, pool, network)
	defer purgeContainer(t, resource)

	go logContainer(t, pool, resource)

	btcdIp := resource.GetIPInNetwork(network)
	rpcAddress := btcdIp + ":8334"

	// Wait for RPC port to be open
	err := pool.Retry(func() error {
		conn, err := net.Dial("tcp", rpcAddress)
		if err != nil {
			t.Logf("Not able to connect to %s, err %s", rpcAddress, err)
		} else {
			conn.Close()
		}
		return err
	})
	if err != nil {
		t.Fatalf("RPC port not open, err %s", err)
	}

	conn, err := tls.Dial("tcp", rpcAddress, &tls.Config{
		InsecureSkipVerify: true, // Do not verify TLS Certificate since it's self signed.
	})
	if err != nil {
		t.Fatal("Failed to connect: " + err.Error())
	}
	conn.Close()
}

func TestRPCConnection(t *testing.T) {
	pool, network := createBtcdNetwork(t)
	defer purgeNetwork(t, network)

	resource := startBtcdWithRPC(t, pool, network)
	defer purgeContainer(t, resource)

	go logContainer(t, pool, resource)

	btcdIp := resource.GetIPInNetwork(network)
	rpcAddress := btcdIp + ":8334"

	// Wait for RPC port to be open
	err := pool.Retry(func() error {
		conn, err := net.Dial("tcp", rpcAddress)
		if err != nil {
			t.Logf("Not able to connect to %s, err %s", rpcAddress, err)
		} else {
			conn.Close()
		}
		return err
	})
	if err != nil {
		t.Fatalf("RPC port not open, err %s", err)
	}

	conn, err := tls.Dial("tcp", rpcAddress, &tls.Config{
		InsecureSkipVerify: true, // Do not verify TLS Certificate since it's self signed.
	})
	if err != nil {
		t.Fatal("Failed to connect: " + err.Error())
	}

	// Store RPC Server certificates so that rpcclient can use it
	var rpcServerCerts bytes.Buffer
	for _, cert := range conn.ConnectionState().PeerCertificates {
		err := pem.Encode(&rpcServerCerts, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})
		if err != nil {
			t.Fatalf("Failed to save to memory RPC Server certificates, err %s", err)
		}
	}
	conn.Close()

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
	client.Shutdown()
}

func TestRPCGetBlockCount(t *testing.T) {
	pool, network := createBtcdNetwork(t)
	defer purgeNetwork(t, network)

	resource := startBtcdWithRPC(t, pool, network)
	defer purgeContainer(t, resource)

	go logContainer(t, pool, resource)

	client := rpcConnect(t, pool, network, resource)
	defer client.Shutdown()

	// Get the current block count.
	_, err := client.GetBlockCount()
	if err != nil {
		t.Fatalf("Failed to perform RPC call GetBlockCount, err %s", err)
	}
}
