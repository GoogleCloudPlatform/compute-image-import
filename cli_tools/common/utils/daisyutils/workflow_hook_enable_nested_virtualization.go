//  Copyright 2022 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package daisyutils

import (
	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

// EnableNestedVirtualizationHook is a WorkflowHook that updates CreateInstances in a
// daisy workflow such that they will be created with nested virtualization enabled.
//
// For more info on nested virtualization see:
//
//	https://cloud.google.com/compute/docs/instances/nested-virtualization/overview
type EnableNestedVirtualizationHook struct{}

// PreRunHook updates the CreateInstances steps so that they won't have an external IP.
func (t *EnableNestedVirtualizationHook) PreRunHook(wf *daisy.Workflow) error {
	wf.IterateWorkflowSteps(func(step *daisy.Step) {
		if step.CreateInstances != nil {
			for _, instance := range step.CreateInstances.Instances {
				if instance.AdvancedMachineFeatures == nil {
					instance.AdvancedMachineFeatures = &compute.AdvancedMachineFeatures{}
				}
				instance.AdvancedMachineFeatures.EnableNestedVirtualization = true
			}
			for _, instance := range step.CreateInstances.InstancesBeta {
				if instance.AdvancedMachineFeatures == nil {
					instance.AdvancedMachineFeatures = &computeBeta.AdvancedMachineFeatures{}
				}
				instance.AdvancedMachineFeatures.EnableNestedVirtualization = true
			}

		}
	})
	return nil
}
