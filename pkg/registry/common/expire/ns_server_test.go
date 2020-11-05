// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expire_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/sdk/pkg/registry/common/expire"
	"github.com/networkservicemesh/sdk/pkg/registry/common/setid"
	"github.com/networkservicemesh/sdk/pkg/registry/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/registry/memory"
)

const testPeriod = time.Millisecond * 50

func TestNewNetworkServiceRegistryServer(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nseMem := next.NewNetworkServiceEndpointRegistryServer(
		setid.NewNetworkServiceEndpointRegistryServer(),
		memory.NewNetworkServiceEndpointRegistryServer(),
	)

	expiration := time.Now().Add(testPeriod * 2)
	_, err := nseMem.Register(ctx, &registry.NetworkServiceEndpoint{
		NetworkServiceNames: []string{"IP terminator"},
		ExpirationTime: &timestamp.Timestamp{
			Seconds: expiration.Unix(),
			Nanos:   int32(expiration.Nanosecond()),
		},
	})
	require.Nil(t, err)
	nseClient := adapters.NetworkServiceEndpointServerToClient(nseMem)
	nsMem := memory.NewNetworkServiceRegistryServer()
	s := next.NewNetworkServiceRegistryServer(expire.NewNetworkServiceServer(ctx, nseClient), nsMem)
	_, err = s.Register(ctx, &registry.NetworkService{
		Name: "IP terminator",
	})
	require.Nil(t, err)
	nsClient := adapters.NetworkServiceServerToClient(s)
	stream, err := nsClient.Find(ctx, &registry.NetworkServiceQuery{
		NetworkService: &registry.NetworkService{},
	})
	require.Nil(t, err)
	list := registry.ReadNetworkServiceList(stream)
	require.NotEmpty(t, list)

	require.Eventually(t, func() bool {
		stream, err = nsClient.Find(ctx, &registry.NetworkServiceQuery{
			NetworkService: &registry.NetworkService{},
		})
		require.Nil(t, err)
		list = registry.ReadNetworkServiceList(stream)
		return len(list) == 0
	}, time.Second, time.Millisecond*100)
}

func TestNewNetworkServiceRegistryServer_NSEUnregister(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nseMem := next.NewNetworkServiceEndpointRegistryServer(
		memory.NewNetworkServiceEndpointRegistryServer(),
	)
	expiration := time.Now().Add(time.Hour)
	_, err := nseMem.Register(ctx, &registry.NetworkServiceEndpoint{
		Name:                "nse-1",
		NetworkServiceNames: []string{"IP terminator"},
		ExpirationTime: &timestamp.Timestamp{
			Seconds: expiration.Unix(),
			Nanos:   int32(expiration.Nanosecond()),
		},
	})
	require.Nil(t, err)
	nseClient := adapters.NetworkServiceEndpointServerToClient(nseMem)
	nsMem := memory.NewNetworkServiceRegistryServer()
	s := next.NewNetworkServiceRegistryServer(expire.NewNetworkServiceServer(ctx, nseClient), nsMem)
	_, err = s.Register(ctx, &registry.NetworkService{
		Name: "IP terminator",
	})
	require.Nil(t, err)
	nsClient := adapters.NetworkServiceServerToClient(s)
	stream, err := nsClient.Find(ctx, &registry.NetworkServiceQuery{
		NetworkService: &registry.NetworkService{},
	})
	require.Nil(t, err)
	list := registry.ReadNetworkServiceList(stream)
	require.NotEmpty(t, list)
	<-time.After(testPeriod * 2)
	_, err = nseClient.Unregister(ctx, &registry.NetworkServiceEndpoint{
		Name:                "nse-1",
		NetworkServiceNames: []string{"IP terminator"},
	})
	require.Nil(t, err)
	require.Eventually(t, func() bool {
		stream, err = nsClient.Find(ctx, &registry.NetworkServiceQuery{
			NetworkService: &registry.NetworkService{},
		})
		require.Nil(t, err)
		list = registry.ReadNetworkServiceList(stream)
		return len(list) == 0
	}, time.Second, time.Millisecond*100)
}
