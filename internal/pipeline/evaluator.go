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

package pipeline

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hacbs-contract/ec-cli/internal/utils"
	"github.com/open-policy-agent/conftest/runner"
	"github.com/spf13/afero"
)

var createWorkDir = afero.TempDir

//Evaluator is a struct containing the required elements to evaluate an associated EvaluationTarget
//using associated PolicySource objects.
type Evaluator struct {
	Context       context.Context
	Target        EvaluationTarget
	PolicySources []PolicySource
	Paths         ConfigurationPaths
	TestRunner    runner.TestRunner
	Namespace     []string
	OutputFormat  string
	workDir       string
}

//ConfigurationPaths is a structs containing necessary paths for an Evaluator struct
type ConfigurationPaths struct {
	PolicyPaths []string
	DataPaths   []string
}

func (e *Evaluator) addDataPath() error {
	dataDir := filepath.Join(e.workDir, "data")
	exists, err := afero.DirExists(utils.AppFS, dataDir)
	if err != nil {
		return err
	}
	if !exists {
		_ = utils.AppFS.MkdirAll(dataDir, 0755)
		err = afero.WriteFile(utils.AppFS, filepath.Join(e.workDir, "data/data.json"), []byte("{\"config\":{}}\n"), 0777)
		if err != nil {
			return err
		}
	}
	e.Paths.DataPaths = append(e.Paths.DataPaths, dataDir)
	return nil
}
func (e *Evaluator) addPolicyPaths() error {
	for i := range e.PolicySources {
		err := e.PolicySources[i].getPolicies(e.workDir)
		if err != nil {
			return err
		}
		policyDir := e.PolicySources[i].getPolicyDir()
		policyPath := filepath.Join(e.workDir, policyDir)
		e.Paths.PolicyPaths = append(e.Paths.PolicyPaths, policyPath)
	}
	return nil
}
func (e *Evaluator) createWorkDir() (string, error) {
	return createWorkDir(utils.AppFS, afero.GetTempDir(utils.AppFS, ""), "ec-work-")
}

//NewPipelineEvaluator returns an *Evaluator specific to Pipeline validation
func NewPipelineEvaluator(ctx context.Context, fpath string, policyRepo PolicyRepo, namespace string) (*Evaluator, error) {
	e := &Evaluator{
		Context:       ctx,
		Target:        &DefinitionFile{fpath: fpath},
		Paths:         ConfigurationPaths{},
		PolicySources: []PolicySource{&policyRepo},
		Namespace:     []string{namespace},
		OutputFormat:  "json",
	}
	exists, err := e.Target.exists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("definition file `%s` does not exist", fpath)
	}
	workDir, err := e.createWorkDir()
	if err != nil {
		return nil, err
	}
	e.workDir = workDir
	err = e.addPolicyPaths()
	if err != nil {
		return nil, err
	}
	err = e.addDataPath()
	if err != nil {
		return nil, err
	}
	e.TestRunner = runner.TestRunner{
		Policy:    e.Paths.PolicyPaths,
		Data:      e.Paths.DataPaths,
		Namespace: e.Namespace,
		NoFail:    true,
		Output:    e.OutputFormat,
	}
	return e, nil
}