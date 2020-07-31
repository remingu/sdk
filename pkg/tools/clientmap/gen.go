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

// Package clientmap provides a sync.Map for networkservicemesh.NetworkServiceClients and a refcounted sync.Map
package clientmap

import (
	"sync"
)

//go:generate go-syncmap -output clientmap.gen.go -type Map<string,github.com/networkservicemesh/api/pkg/api/networkservice.NetworkServiceClient>

// Map - sync.Map with key == string and value == networkservice.NetworkServiceClient
type Map sync.Map
