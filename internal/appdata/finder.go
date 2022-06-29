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
	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	triggers "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"sigs.k8s.io/yaml"
)

type K8sObject struct {
	Kind string `json:"kind,omitempty"`
}

func findBundles(raw []byte) (bundles []string, err error) {
	var object K8sObject
	// TODO: What if the file contains more than one yaml document? Check all
	err = yaml.Unmarshal(raw, &object)
	if err != nil {
		return
	}

	switch object.Kind {
	case "TriggerTemplate":
		var triggerTemplate triggers.TriggerTemplate
		err = yaml.Unmarshal(raw, &triggerTemplate)
		if err != nil {
			return
		}
		bundles, err = findBundlesInTriggerTemplate(triggerTemplate)
	}

	return
}

func findBundlesInTriggerTemplate(tt triggers.TriggerTemplate) (bundles []string, err error) {
	for _, resourceTemplate := range tt.Spec.ResourceTemplates {
		var object K8sObject
		err = yaml.Unmarshal(resourceTemplate.Raw, &object)
		if err != nil {
			return
		}
		switch object.Kind {
		case "PipelineRun":
			var pipelineRun pipeline.PipelineRun
			err = yaml.Unmarshal(resourceTemplate.Raw, &pipelineRun)
			if err != nil {
				return
			}
			if pipelineRun.Spec.PipelineRef == nil {
				continue
			}
			bundle := findBundleInPipelineRun(pipelineRun)
			if bundle == "" {
				continue
			}
			bundles = append(bundles, bundle)
		}
	}

	return
}

func findBundleInPipelineRun(pr pipeline.PipelineRun) (bundle string) {
	if pr.Spec.PipelineRef == nil {
		return
	}
	bundle = pr.Spec.PipelineRef.Bundle
	return
}
