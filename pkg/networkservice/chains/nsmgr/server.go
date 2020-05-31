// Copyright (c) 2020 Cisco and/or its affiliates.
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

// Package nsmgr provides a Network Service Manager (nsmgr), but interface and implementation
package nsmgr

import (
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/client"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/discover"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/localbypass"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/roundrobin"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/adapters"
	adapter_registry "github.com/networkservicemesh/sdk/pkg/registry/core/adapters"
	chain_registry "github.com/networkservicemesh/sdk/pkg/registry/core/chain"
	"github.com/networkservicemesh/sdk/pkg/tools/addressof"
	"github.com/networkservicemesh/sdk/pkg/tools/token"
	"github.com/open-policy-agent/opa/rego"
	"google.golang.org/grpc"
	"net/url"
)

// Nsmgr - A simple combintation of the Endpoint, registry.NetworkServiceRegistryServer, and registry.NetworkServiceDiscoveryServer interfaces
type Nsmgr interface {
	endpoint.Endpoint
	registry.NetworkServiceRegistryServer
	registry.NetworkServiceDiscoveryServer
}

type nsmgr struct {
	endpoint.Endpoint
	registry.NetworkServiceRegistryServer
	registry.NetworkServiceDiscoveryServer
}

//type urlInjector struct {
//	name        string
//	listenOnURL url.URL
//}
//
//func (u *urlInjector) RegisterNSE(ctx context.Context, registration *registry.NSERegistration) (*registry.NSERegistration, error) {
//	expiration := time.Now()
//	expiration.Add(24 * time.Hour)
//	expirationTimestamp, _ := ptypes.TimestampProto(expiration)
//	registration.NetworkServiceManager = &registry.NetworkServiceManager{
//		Name:                 u.name,
//		Url:                  u.listenOnURL.String(),
//		ExpirationTime:       expirationTimestamp,
//		State:                "RUNNING",
//	}
//	return registration, nil
//}
//
//func (u *urlInjector) BulkRegisterNSE(server registry.NetworkServiceRegistry_BulkRegisterNSEServer) error {
//	panic("implement me")
//}
//
//func (u *urlInjector) RemoveNSE(ctx context.Context, request *registry.RemoveNSERequest) (*empty.Empty, error) {
//	panic("implement me")
//}

// NewServer - Creates a new Nsmgr
//           name - name of the Nsmgr
//           authzPolicy - authorization policy
//           registryCC - client connection to reach the upstream registry
func NewServer(name string, authzPolicy *rego.PreparedEvalQuery, tokenGenerator token.GeneratorFunc, registryCC grpc.ClientConnInterface, listenOnURL url.URL) Nsmgr {
	rv := &nsmgr{}
	rv.Endpoint = endpoint.NewServer(
		name,
		authzPolicy,
		tokenGenerator,
		discover.NewServer(registry.NewNetworkServiceDiscoveryClient(registryCC)),
		roundrobin.NewServer(),
		localbypass.NewServer(&rv.NetworkServiceRegistryServer),
		connect.NewServer(client.NewClientFactory(name, addressof.NetworkServiceClient(adapters.NewServerToClient(rv)), tokenGenerator)),
	)
	rv.NetworkServiceRegistryServer = chain_registry.NewNetworkServiceRegistryServer(
		rv.NetworkServiceRegistryServer,
		//&urlInjector{
		//	name:        "nsmgr",
		//	listenOnURL: listenOnURL,
		//},
		adapter_registry.NewRegistryClientToServer(registry.NewNetworkServiceRegistryClient(registryCC)),
	)
	rv.NetworkServiceDiscoveryServer = adapter_registry.NewDiscoveryClientToServer(registry.NewNetworkServiceDiscoveryClient(registryCC))
	return rv
}

func (n *nsmgr) Register(s *grpc.Server) {
	n.Endpoint.Register(s)
	registry.RegisterNetworkServiceRegistryServer(s, n)
	registry.RegisterNetworkServiceDiscoveryServer(s, n)
}
