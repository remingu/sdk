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

// Package nextwrap provides adapters to wrap clients with no support to next, to support it.
package nextwrap

import (
	"context"
	"io"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
)

type nextNetworkServiceWrappedClient struct {
	client registry.NetworkServiceRegistryClient
}

func (r *nextNetworkServiceWrappedClient) Register(ctx context.Context, in *registry.NetworkService, opts ...grpc.CallOption) (*registry.NetworkService, error) {
	result, err := r.client.Register(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return next.NetworkServiceRegistryClient(ctx).Register(ctx, result, opts...)
}

func (r *nextNetworkServiceWrappedClient) Find(ctx context.Context, in *registry.NetworkServiceQuery, opts ...grpc.CallOption) (registry.NetworkServiceRegistry_FindClient, error) {
	client, err := r.client.Find(ctx, in, opts...)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if client != nil {
		return client, nil
	}
	return next.NetworkServiceRegistryClient(ctx).Find(ctx, in, opts...)
}

func (r *nextNetworkServiceWrappedClient) Unregister(ctx context.Context, in *registry.NetworkService, opts ...grpc.CallOption) (*empty.Empty, error) {
	_, err := r.client.Unregister(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return next.NetworkServiceRegistryClient(ctx).Unregister(ctx, in, opts...)
}

// NewNetworkServiceRegistryClient wraps NetworkServiceRegistryClient to support next chain elements
func NewNetworkServiceRegistryClient(client registry.NetworkServiceRegistryClient) registry.NetworkServiceRegistryClient {
	return &nextNetworkServiceWrappedClient{client: client}
}

var _ registry.NetworkServiceRegistryClient = &nextNetworkServiceWrappedClient{}
