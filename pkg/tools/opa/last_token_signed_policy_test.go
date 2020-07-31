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

package opa_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/networkservicemesh/sdk/pkg/tools/opa"
)

func TestVerifyToken(t *testing.T) {
	ca, err := generateCA()
	require.Nil(t, err)

	cert, err := generateKeyPair(spiffeID, "test.com", &ca)
	require.Nil(t, err)

	token, err := jwt.New(jwt.SigningMethodES256).SignedString(cert.PrivateKey)
	require.Nil(t, err)

	x509crt, err := x509.ParseCertificate(cert.Certificate[0])
	require.Nil(t, err)

	p := opa.WithLastTokenSignedPolicy()

	input := networkservice.Path{
		PathSegments: []*networkservice.PathSegment{
			{
				Token: token,
			},
		},
	}

	peerAuth := &peer.Peer{
		AuthInfo: &credentials.TLSInfo{
			State: tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					x509crt,
				},
			},
		},
	}

	ctx := peer.NewContext(context.Background(), peerAuth)

	err = p.Check(ctx, input)
	require.Nil(t, err)

	invalidX509crt, err := x509.ParseCertificate(ca.Certificate[0])
	require.Nil(t, err)

	peerAuth.AuthInfo.(*credentials.TLSInfo).State.PeerCertificates[0] = invalidX509crt

	err = p.Check(ctx, input)
	require.NotNil(t, err)
}
