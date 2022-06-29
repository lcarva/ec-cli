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

package cmd

import (
	"context"
	"fmt"

	"github.com/hacbs-contract/ec-cli/internal/appdata"
	"github.com/spf13/cobra"
)

type appdataValidationFn func(context.Context, string) error

func validateAppdataCmd(validate appdataValidationFn) *cobra.Command {
	var args = struct {
		RepoDir string
	}{
		RepoDir: ".",
	}
	cmd := &cobra.Command{
		Use:     "appdata",
		Short:   "Validates an appdata git repo",
		Long:    `Validate the Tekton resources defined in, or referenced by, an appdata git repository.`,
		Example: "",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Printf("Processing appdata git repo in %s\n", args.RepoDir)
			err := appdata.Validate(cmd.Context(), args.RepoDir)
			return err
		},
	}
	cmd.Flags().StringVar(&args.RepoDir, "repo-dir", args.RepoDir, "directory containing the cloned appdata git repo")
	return cmd
}
