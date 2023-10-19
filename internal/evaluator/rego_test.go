// Copyright The Enterprise Contract Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unit

package evaluator

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/enterprise-contract/ec-cli/internal/fetchers/oci"
	"github.com/enterprise-contract/ec-cli/internal/fetchers/oci/fake"
)

func TestOCIBlob(t *testing.T) {
	cases := []struct {
		name      string
		data      string
		uri       *ast.Term
		err       bool
		remoteErr error
	}{
		{
			name: "success",
			data: `{"spam": "maps"}`,
			uri:  ast.StringTerm("registry.local/spam@sha256:4bbf56a3a9231f752d3b9c174637975f0f83ed2b15e65799837c571e4ef3374b"),
		},
		{
			name: "unexpected uri type",
			data: `{"spam": "maps"}`,
			uri:  ast.IntNumberTerm(42),
			err:  true,
		},
		{
			name: "missing digest",
			data: `{"spam": "maps"}`,
			uri:  ast.StringTerm("registry.local/spam:latest"),
			err:  true,
		},
		{
			name: "invalid digest size",
			data: `{"spam": "maps"}`,
			uri:  ast.StringTerm("registry.local/spam@sha256:4e388ab"),
			err:  true,
		},
		{
			name:      "remote error",
			data:      `{"spam": "maps"}`,
			uri:       ast.StringTerm("registry.local/spam@sha256:4bbf56a3a9231f752d3b9c174637975f0f83ed2b15e65799837c571e4ef3374b"),
			remoteErr: errors.New("boom!"),
			err:       true,
		},
		{
			name: "unexpected digest",
			data: `{"spam": "mapssssss"}`,
			uri:  ast.StringTerm("registry.local/spam@sha256:4bbf56a3a9231f752d3b9c174637975f0f83ed2b15e65799837c571e4ef3374b"),
			err:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := fake.FakeClient{}
			if c.remoteErr != nil {
				client.On("Layer", mock.Anything, mock.Anything).Return(nil, c.remoteErr)
			} else {
				layer := static.NewLayer([]byte(c.data), specs.MediaTypeImageLayer)
				client.On("Layer", mock.Anything, mock.Anything).Return(layer, nil)
			}
			ctx := oci.WithClient(context.Background(), &client)
			bctx := rego.BuiltinContext{Context: ctx}

			blob, err := ociBlob(bctx, c.uri)
			require.NoError(t, err)
			if c.err {
				require.Nil(t, blob)
			} else {
				require.NotNil(t, blob)
				data, ok := blob.Value.(ast.String)
				require.True(t, ok)
				require.Equal(t, c.data, string(data))
			}
		})
	}
}

func TestFunctionsRegistered(t *testing.T) {
	names := []string{
		ociBlobName,
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			for _, builtin := range ast.Builtins {
				if builtin.Name == name {
					return
				}
			}
			t.Fatalf("%s builtin not registered", name)
		})
	}
}
