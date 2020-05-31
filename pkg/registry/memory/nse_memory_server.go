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

package memory

import (
	"context"
	"github.com/networkservicemesh/sdk/pkg/tools/log"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
)

type memoryNetworkServeRegistry struct {
	storage *Storage
	nsmName string
}

func (m *memoryNetworkServeRegistry) RegisterNSE(ctx context.Context, registration *registry.NSERegistration) (*registry.NSERegistration, error) {
	log.Entry(ctx).Println("Register NSE: %+v", registration)
	if registration == nil {
		return nil, errors.New("can not register nil registration")
	}
	registration.NetworkServiceEndpoint.State = "RUNNING"
	registration.NetworkServiceEndpoint.NetworkServiceManagerName = m.nsmName
	nsm, ok := m.storage.NetworkServiceManagers.Load(m.nsmName)
	if !ok {
		return nil, errors.New("network service manager is not found")
	}
	registration.NetworkServiceManager = nsm

	m.storage.NetworkServiceEndpoints.Store(registration.NetworkServiceEndpoint.Name, registration.NetworkServiceEndpoint)
	m.storage.NetworkServices.Store(registration.NetworkService.Name, registration.NetworkService)
	return next.NetworkServiceRegistryServer(ctx).RegisterNSE(ctx, registration)
}

func (m *memoryNetworkServeRegistry) BulkRegisterNSE(s registry.NetworkServiceRegistry_BulkRegisterNSEServer) error {
	return next.NetworkServiceRegistryServer(s.Context()).BulkRegisterNSE(s)
}

func (m *memoryNetworkServeRegistry) RemoveNSE(ctx context.Context, req *registry.RemoveNSERequest) (*empty.Empty, error) {
	log.Entry(ctx).Println("Remove NSE: %+v", req)
	m.storage.NetworkServiceEndpoints.Delete(req.NetworkServiceEndpointName)
	return next.NetworkServiceRegistryServer(ctx).RemoveNSE(ctx, req)
}

// NewNetworkServiceRegistryServer returns new NetworkServiceRegistryServer based on specific resource client
func NewNetworkServiceRegistryServer(nsmName string, storage *Storage) registry.NetworkServiceRegistryServer {
	return &memoryNetworkServeRegistry{
		nsmName: nsmName,
		storage: storage,
	}
}

var _ registry.NetworkServiceRegistryServer = &memoryNetworkServeRegistry{}
