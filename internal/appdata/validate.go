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

package appdata

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Validate(ctx context.Context, repoDir string) error {
	ignorePrefix := path.Join(repoDir, ".git") + string(os.PathSeparator)

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(path, ignorePrefix) {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		bundles, err := findBundles(content)
		if err != nil {
			return err
		}
		if len(bundles) > 0 {
			// TODO: Remove this Println
			fmt.Println(path, "contains bundles:", bundles)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
