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

package connect_test

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/refresh"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/setextracontext"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/updatepath"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/updatetoken"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/inject/injectpeer"
	"github.com/networkservicemesh/sdk/pkg/tools/clienturlctx"
	"github.com/networkservicemesh/sdk/pkg/tools/sandbox"
	"github.com/networkservicemesh/sdk/pkg/tools/token"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

const (
	timeout = 10 * time.Second
)

type nseTest struct {
	ctx      context.Context
	cancel   context.CancelFunc
	listenOn *url.URL
	t        *testing.T
	nse      endpoint.Endpoint
	errCh    <-chan error
}

func (nseT *nseTest) Stop() {
	nseT.cancel()
	// try read value from err channel, with this we will wait for cancel to be processed and all go routines will exit
	<-nseT.errCh
}

func (nseT *nseTest) Setup() {
	nseT.ctx, nseT.cancel = context.WithTimeout(context.Background(), 50*time.Second)
	nseT.listenOn = &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	nseT.nse = endpoint.NewServer(nseT.ctx, "testServer", authorize.NewServer(),
		sandbox.GenerateTestToken,
		setextracontext.NewServer(map[string]string{"ok": "all is ok"}))

	nseT.errCh = endpoint.Serve(nseT.ctx, nseT.listenOn, nseT.nse)
}

func (nseT *nseTest) newNSEContext(ctx context.Context) context.Context {
	return clienturlctx.WithClientURL(ctx, &url.URL{Scheme: "tcp", Host: nseT.listenOn.Host})
}

func newClient(ctx context.Context, generatorFunc token.GeneratorFunc) networkservice.NetworkServiceClient {
	return chain.NewNetworkServiceClient(
		append([]networkservice.NetworkServiceClient{
			authorize.NewClient(),
			updatepath.NewClient("nsc-" + uuid.New().String()),
			refresh.NewClient(ctx),
		},
			injectpeer.NewClient(),
			updatetoken.NewClient(generatorFunc),
			adapters.NewServerToClient(connect.NewServer(ctx, func(ctx context.Context, cc grpc.ClientConnInterface) networkservice.NetworkServiceClient {
				return networkservice.NewNetworkServiceClient(cc)
			}, grpc.WithBlock(), grpc.WithInsecure())),
		)...)
}

func TestConnectServerShouldNotPanicOnRequest(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	nseT := &nseTest{
		t: t,
	}
	nseT.Setup()
	defer nseT.Stop()

	t.Run("Check Request", func(t *testing.T) {
		require.NotPanics(t, func() {
			clientURLCtx, clientCancel := context.WithTimeout(context.Background(), timeout)
			defer clientCancel()
			clientURLCtx = nseT.newNSEContext(clientURLCtx)

			c := newClient(clientURLCtx, sandbox.GenerateTestToken)
			conn, err := c.Request(clientURLCtx, &networkservice.NetworkServiceRequest{
				Connection: &networkservice.Connection{
					Id: "1",
				},
			})
			require.Nil(t, err)
			require.NotNil(t, conn)
			require.Equal(t, "all is ok", conn.GetContext().GetExtraContext()["ok"])
			_, err = c.Close(clientURLCtx, &networkservice.Connection{
				Id: conn.Id,
			})
			require.Nil(t, err)
		})
	})
	t.Run("Close Id", func(t *testing.T) {
		require.NotPanics(t, func() {
			clientURLCtx, clientCancel := context.WithTimeout(context.Background(), timeout)
			defer clientCancel()
			clientURLCtx = nseT.newNSEContext(clientURLCtx)

			c := newClient(clientURLCtx, sandbox.GenerateTestToken)
			conn, err := c.Request(clientURLCtx, &networkservice.NetworkServiceRequest{
				Connection: &networkservice.Connection{
					Id: "1",
				},
			})
			require.Nil(t, err)
			require.NotNil(t, conn)
			require.Equal(t, "all is ok", conn.GetContext().GetExtraContext()["ok"])

			// Do not pass clientURL
			_, err = c.Close(context.Background(), &networkservice.Connection{
				Id: "1",
			})
			require.Nil(t, err)
		})
	})
	t.Run("Check no clientURL", func(t *testing.T) {
		require.NotPanics(t, func() {
			clientURLCtx, clientCancel := context.WithTimeout(context.Background(), timeout)
			defer clientCancel()

			c := newClient(clientURLCtx, sandbox.GenerateTestToken)
			conn, err := c.Request(context.Background(), &networkservice.NetworkServiceRequest{
				Connection: &networkservice.Connection{
					Id: "1",
				},
			})
			require.NotNil(t, err)
			require.Nil(t, conn)
		})
	})
	t.Run("Request without client URL", func(t *testing.T) {
		require.NotPanics(t, func() {
			clientURLCtx, clientCancel := context.WithTimeout(context.Background(), timeout)
			defer clientCancel()
			clientURLCtx = nseT.newNSEContext(clientURLCtx)

			c := newClient(clientURLCtx, sandbox.GenerateTestToken)
			conn, err := c.Request(clientURLCtx, &networkservice.NetworkServiceRequest{
				Connection: &networkservice.Connection{
					Id: "1",
				},
			})
			require.Nil(t, err)
			require.NotNil(t, conn)
			require.Equal(t, "all is ok", conn.GetContext().GetExtraContext()["ok"])

			// Request again
			conn, err = c.Request(context.Background(), &networkservice.NetworkServiceRequest{
				Connection: &networkservice.Connection{
					Id: "1",
				},
			})
			require.Nil(t, err)
			require.NotNil(t, conn)
			require.Equal(t, "all is ok", conn.GetContext().GetExtraContext()["ok"])

			// Do not pass clientURL
			_, err = c.Close(context.Background(), &networkservice.Connection{
				Id: "1",
			})
			require.Nil(t, err)
		})
	})
}

func TestParallelDial(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	nseT := &nseTest{}
	nseT.Setup()
	defer nseT.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		j := i
		clientURLCtx, clientCancel := context.WithTimeout(context.Background(), timeout)
		defer clientCancel()
		clientURLCtx = nseT.newNSEContext(clientURLCtx)
		go func() {
			defer wg.Done()
			for k := 0; k < 10; k++ {
				c := newClient(clientURLCtx, sandbox.GenerateTestToken)
				conn, err := c.Request(clientURLCtx, &networkservice.NetworkServiceRequest{
					Connection: &networkservice.Connection{
						Id: fmt.Sprintf("%d", j),
					},
				})
				require.Nil(t, err)
				require.NotNil(t, conn)

				_, err = c.Close(clientURLCtx, conn)

				require.Nil(t, err)
			}
		}()
	}
	wg.Wait()
}
