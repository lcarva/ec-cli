// Copyright 2022 Red Hat, Inc.
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

package application_snapshot_image

import (
	"context"

	"github.com/google/go-containerregistry/pkg/name"
	gcr "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type contextKey string

const clientContextKey contextKey = "ec.appliation-snapshot-image.client"

// Client is an interface that contains all the external calls used by the
// application_snapshot_image package.
type Client interface {
	Head(name.Reference, ...remote.Option) (*gcr.Descriptor, error)
}

func WithClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, clientContextKey, client)
}

// NewClient constructs a new application_snapshot_image with the default client.
func NewClient(ctx context.Context) Client {
	client, ok := ctx.Value(clientContextKey).(Client)
	if ok && client != nil {
		return client
	}

	return &defaultClient{}
}

type defaultClient struct {
}

func (c *defaultClient) Head(ref name.Reference, opts ...remote.Option) (*gcr.Descriptor, error) {
	return remote.Head(ref, opts...)
}
