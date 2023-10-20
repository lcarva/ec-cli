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

package oci

import (
	"context"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	log "github.com/sirupsen/logrus"
)

type key string

const clientKey key = "ec.fetcher.config.client"

var imgCache cache.Cache

func init() {
	initCache()
}

func initCache() {
	// if a value was set and it is parsed as false, turn the cache off
	if v, err := strconv.ParseBool(os.Getenv("EC_CACHE")); err == nil && !v {
		return
	}

	if userCache, err := os.UserCacheDir(); err != nil {
		log.Debug("unable to find user cache directory")
	} else {
		imgCacheDir := path.Join(userCache, "ec", "images")
		if err := os.MkdirAll(imgCacheDir, 0700); err != nil {
			log.Debugf("unable to create temporary directory for image cache in %q: %v", imgCacheDir, err)
		}
		log.Debugf("using %q directory to store image cache", imgCacheDir)
		imgCache = cache.NewFilesystemCache(imgCacheDir)
	}
}

type client interface {
	Image(name.Reference, ...remote.Option) (v1.Image, error)
	Layer(name.Digest, ...remote.Option) (v1.Layer, error)
}

var defaultClient = remoteClient{}

func NewClient(ctx context.Context) client {
	c, ok := ctx.Value(clientKey).(client)
	if ok && c != nil {
		return c
	}

	return &defaultClient
}

func WithClient(ctx context.Context, c client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

type remoteClient struct {
}

func (*remoteClient) Image(ref name.Reference, opts ...remote.Option) (v1.Image, error) {
	img, err := remote.Image(ref, opts...)
	if err != nil {
		return nil, err
	}

	if imgCache != nil {
		img = cache.Image(img, imgCache)
	}

	return img, nil
}

// TODO: Wrap all the returned errors
func (*remoteClient) Layer(ref name.Digest, opts ...remote.Option) (v1.Layer, error) {
	if imgCache == nil {
		layer, err := remote.Layer(ref, opts...)
		if err != nil {
			return nil, err
		}
		return layer, nil
	}

	hash := v1.Hash{}
	hash.Algorithm, hash.Hex, _ = strings.Cut(ref.DigestStr(), ":")
	layer, err := imgCache.Get(hash)
	if err != nil {
		if err != cache.ErrNotFound {
			return nil, err
		}
		// TODO: hmmm this fetches the layer from the registry. But then it is fetched again
		// later when using the cache. Each layer is always fetched exactly 2 times, regardless
		// of how many times this function is called, and that's bad....
		layer, err := remote.Layer(ref, opts...)
		if err != nil {
			return nil, err
		}
		cachedLayer, err := imgCache.Put(layer)
		if err != nil {
			return nil, err
		}
		return cachedLayer, nil
	}
	return layer, nil

	// manifest := `
	// `
	// imp :=
	// The gcr.cache package doesn't provide an API to cache layers directly. To work around
	// this, we wrap the layer in an image and cache that.
	// manifest, err := empty.Image.Manifest()
	// if err != nil {
	// 	return nil, err
	// }

	// descriptor, err := remote.Head(ref, opts...)
	// if err != nil {
	// 	return nil, err
	// }
	// manifest.Layers = append(manifest.Layers, *descriptor)

	// // yolo := v1.Descriptor{
	// // 	MediaType:    specs.MediaTypeImageLayer,
	// // 	Size:         0,
	// // 	Digest:       v1.Hash{},
	// // 	Data:         []byte{},
	// // 	URLs:         []string{},
	// // 	Annotations:  map[string]string{},
	// // 	Platform:     &v1.Platform{},
	// // 	ArtifactType: "",
	// // }
	// // Layers:[]v1.Descriptor{v1.Descriptor{MediaType:"application/vnd.docker.image.rootfs.diff.tar.gzip", Size:4352, Digest:v1.Hash{Algorithm:"sha256", Hex:"1ff7945e970d4ed01d8942ac5e7ca268bed25b047707cfcef2f9cf8e0cb1b3f7"}

	// // img, err := mutate.AppendLayers(empty.Image, descriptor)
	// // if err != nil {
	// // 	return nil, err
	// // }

	// layers, err := img.Layers()
	// if err != nil {
	// 	return nil, err
	// }

	// if len(layers) != 1 {
	// 	return nil, fmt.Errorf("unexpected amount of layers, %d", len(layers))
	// }

	// return nil, nil
	// layer = layers[0]
	// return layer, nil
}
