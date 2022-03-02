// Copyright (c) 2022 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dockertests

import (
	"testing"
)

func TestRPCNotListeningByDefault(t *testing.T) {
	pool, resource := startBtcd(t)
	go logContainer(t, pool, resource)
	defer purgeContainer(t, pool, resource)

	rpcListeningPort := resource.GetPort("8334/tcp")
	if rpcListeningPort != "" {
		t.Fatalf("Unexpected exposed rpc port. Listening port %v", rpcListeningPort)
	}
}
