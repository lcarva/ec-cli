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

package evaluator

import (
	"bytes"
	"io"
	"strconv"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/types"
)

func init() {
	registerFetchOCIBlob()
}

// ec.fetch_oci_blob

func registerFetchOCIBlob() {
	decl := rego.Function{
		Name: "ec.fetch_oci_blob",
		Decl: types.NewFunction(
			types.Args(
				types.Named("ref", types.S).Description("OCI blob reference"),
			),
			types.Named("blob", types.S).Description("the OCI blob"),
		),
	}

	rego.RegisterBuiltin1(&decl, fetchOCIBlob)
}

const maxBytes = 10 * 1024 * 1024 // 10 MB

func fetchOCIBlob(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
	p, err := builtins.StringOperand(op1.Value, 1)
	if err != nil {
		return nil, err
	}

	// TODO: what's up with this?! This shouldn't be needed...
	uri, err := strconv.Unquote(p.String())
	if err != nil {
		return nil, err
	}

	ref, err := name.NewDigest(uri)
	if err != nil {
		return nil, err
	}

	rawLayer, err := remote.Layer(ref)
	if err != nil {
		return nil, err
	}

	layer, err := rawLayer.Compressed()
	if err != nil {
		return nil, err
	}
	defer layer.Close()

	var blob bytes.Buffer
	if _, err := io.Copy(&blob, io.LimitReader(layer, maxBytes)); err != nil {
		return nil, err
	}

	return ast.StringTerm(blob.String()), nil
}
