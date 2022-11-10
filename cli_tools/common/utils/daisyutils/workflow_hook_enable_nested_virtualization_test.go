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
	"testing"

	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	"github.com/stretchr/testify/assert"
	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

func Test_EnableNestedVirtualizatoin(t *testing.T) {
	w := createWorkflowForNestedVirtualizationTest()
	assert.NoError(t, (&EnableNestedVirtualizationHook{}).PreRunHook(w))

	assert.True(t, (*w.Steps["ci"].CreateInstances).Instances[0].Instance.AdvancedMachineFeatures.EnableNestedVirtualization)
	assert.True(t, (*w.Steps["ci"].CreateInstances).InstancesBeta[0].Instance.AdvancedMachineFeatures.EnableNestedVirtualization)
}

func createWorkflowForNestedVirtualizationTest() *daisy.Workflow {
	w := daisy.New()
	w.Steps = map[string]*daisy.Step{
		"ci": {
			CreateInstances: &daisy.CreateInstances{
				Instances: []*daisy.Instance{
					{
						Instance: compute.Instance{
							Disks: []*compute.AttachedDisk{{Source: "key1"}},
						},
					},
				},
				InstancesBeta: []*daisy.InstanceBeta{
					{
						Instance: computeBeta.Instance{
							Disks: []*computeBeta.AttachedDisk{{Source: "key1"}},
						},
					},
				},
			},
		},
	}
	return w
}
