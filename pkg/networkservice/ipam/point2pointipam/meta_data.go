// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

package point2pointipam

import (
	"context"

	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/metadata"
)

type keyType struct{}

func storeConnInfo(ctx context.Context, connInfo *connectionInfo) {
	metadata.Map(ctx, false).Store(keyType{}, connInfo)
}

func loadConnInfo(ctx context.Context) (*connectionInfo, bool) {
	if raw, ok := metadata.Map(ctx, false).Load(keyType{}); ok {
		return raw.(*connectionInfo), true
	}
	return nil, false
}
