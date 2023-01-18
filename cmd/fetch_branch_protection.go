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

	hd "github.com/MakeNowJust/heredoc"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/log"
	"github.com/spf13/cobra"
)

func fetchBranchProtectionCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "branch-protection <repository> <branch> [<branch> ...]",
		Short:   "Fetch branch-protection information for git branch",
		Long:    hd.Doc(`TODO`),
		Example: hd.Doc(`TODO`),
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repository := args[0]
			branches := args[1:]
			data, err := getBranchProtection(cmd.Context(), repository, branches)
			if err != nil {
				return err
			}

			r := report{
				Repository: repository,
				Pass:       true,
			}

			for _, branch := range data.Branches {
				b := reportBranch{
					Name:               *branch.Name,
					HasProtectionRules: *branch.Protected,
				}
				if *branch.Protected {
					b.RestrictForcePush = !*branch.BranchProtectionRule.AllowForcePushes
					b.RestrictDeletion = !*branch.BranchProtectionRule.AllowDeletions
					b.RestrictAdminBypass = *branch.BranchProtectionRule.EnforceAdmins
				}
				r.Pass = r.Pass && b.RestrictForcePush && b.RestrictDeletion && b.RestrictAdminBypass
				r.Branches = append(r.Branches, b)
			}

			out, err := yaml.Marshal(r)
			if err != nil {
				return err
			}
			if _, err := cmd.OutOrStdout().Write(out); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}

func getBranchProtection(ctx context.Context, repository string, branches []string) (checker.BranchProtectionsData, error) {

	rawData := checker.BranchProtectionsData{}

	repo, err := githubrepo.MakeGithubRepo(repository)
	if err != nil {
		return rawData, err
	}

	logger := log.NewLogger(log.InfoLevel)
	repoClient := githubrepo.CreateGithubRepoClient(ctx, logger)

	if err := repoClient.InitRepo(repo, "HEAD", 1); err != nil {
		return rawData, err
	}
	defer repoClient.Close()

	var allErrors error

	for _, branch := range branches {
		branchRef, err := repoClient.GetBranch(branch)
		if err != nil {
			allErrors = multierror.Append(allErrors, err)
			continue
		}
		if branchRef == nil || branchRef.Name == nil || *branchRef.Name == "" {
			allErrors = multierror.Append(allErrors, fmt.Errorf("branch %q does not exist", branch))
			continue
		}

		rawData.Branches = append(rawData.Branches, *branchRef)
	}

	return rawData, allErrors
}

type report struct {
	Repository string         `json:"repository"`
	Pass       bool           `json:"pass"`
	Branches   []reportBranch `json:"branches"`
}

type reportBranch struct {
	Name                string `json:"name"`
	HasProtectionRules  bool   `json:"hasProtectionRules"`
	RestrictForcePush   bool   `json:"restrictForcePush"`
	RestrictDeletion    bool   `json:"restrictDeletion"`
	RestrictAdminBypass bool   `json:"restrictAdminBypass"`
}
