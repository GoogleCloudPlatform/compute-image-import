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
	"errors"
	"testing"

	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	"github.com/stretchr/testify/assert"
	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/logging"
)

var errN2CpusQuota = errors.New("Quota N2_CPUS exceeded")
var errNotN2CpusQuota = errors.New("failed to start workflow")

func Test_UpdateMachineTypesHook_PostRunHook_WhenQuotaErrorAndSecondaryIsSpecified_ThenRetry(t *testing.T) {
	hook := UpdateMachineTypesHook{
		logger:                 logging.NewToolLogger("test"),
		primaryMachineSeries:   "n2",
		secondaryMachineSeries: "n1",
	}
	wantRetry, wrapped := hook.PostRunHook(errN2CpusQuota)
	assert.Equal(t, errN2CpusQuota, wrapped)
	assert.True(t, wantRetry)
	assert.True(t, hook.shouldFallback)
}

func Test_UpdateMachineTypesHook_PostRunHook_WhenQuotaErrorAndSecondaryIsNotSpecified_ThenNoRetry(t *testing.T) {
	hook := UpdateMachineTypesHook{
		logger:               logging.NewToolLogger("test"),
		primaryMachineSeries: "n2",
	}
	wantRetry, wrapped := hook.PostRunHook(errN2CpusQuota)
	assert.Equal(t, errN2CpusQuota, wrapped)
	assert.False(t, wantRetry)
	assert.False(t, hook.shouldFallback)
}

func Test_UpdateMachineTypesHook_PostRunHook_WhenNotQuotaError_ThenNoRetry(t *testing.T) {
	hook := UpdateMachineTypesHook{
		shouldFallback:         true,
		logger:                 logging.NewToolLogger("test"),
		primaryMachineSeries:   "n2",
		secondaryMachineSeries: "n1",
	}
	wantRetry, wrapped := hook.PostRunHook(errNotN2CpusQuota)
	assert.Equal(t, errNotN2CpusQuota, wrapped)
	assert.False(t, wantRetry)
}

func Test_UpdateMachineTypesHook_PostRunHook_OnlyFallsBackOnce(t *testing.T) {
	hook := UpdateMachineTypesHook{
		shouldFallback:         true,
		logger:                 logging.NewToolLogger("test"),
		primaryMachineSeries:   "n2",
		secondaryMachineSeries: "n1",
	}
	wantRetry, wrapped := hook.PostRunHook(errN2CpusQuota)
	assert.Equal(t, errN2CpusQuota, wrapped)
	assert.False(t, wantRetry)
}

func Test_UpdateMachineTypesHook_PreRunHook_WhenFirstRun_UpdatesWorkflowMachineSeriesToPrimary(t *testing.T) {
	hook := UpdateMachineTypesHook{
		logger:               logging.NewToolLogger("test"),
		primaryMachineSeries: "n2",
	}

	wf := createUpdateMachineTypesTestWorkflow()
	assert.NoError(t, hook.PreRunHook(wf))

	assert.Equal(t, "n2-standard-2", (*wf.Steps["ci"].CreateInstances).Instances[0].MachineType)
	assert.Equal(t, "n2-standard-2", (*wf.Steps["ci"].CreateInstances).InstancesBeta[0].MachineType)
}

func Test_UpdateMachineTypesHook_PreRunHook_WhenShouldFallback_UpdatesWorkflowMachineSeriesToSecondary(t *testing.T) {
	hook := UpdateMachineTypesHook{
		logger:                 logging.NewToolLogger("test"),
		primaryMachineSeries:   "n2",
		secondaryMachineSeries: "n1",
		shouldFallback:         true,
	}

	wf := createUpdateMachineTypesTestWorkflow()
	assert.NoError(t, hook.PreRunHook(wf))

	assert.Equal(t, "n1-standard-2", (*wf.Steps["ci"].CreateInstances).Instances[0].MachineType)
	assert.Equal(t, "n1-standard-2", (*wf.Steps["ci"].CreateInstances).InstancesBeta[0].MachineType)
}

func createUpdateMachineTypesTestWorkflow() *daisy.Workflow {
	w := daisy.New()
	w.Steps = map[string]*daisy.Step{
		"ci": {
			CreateInstances: &daisy.CreateInstances{
				Instances: []*daisy.Instance{
					{
						Instance: compute.Instance{
							MachineType: "e2-standard-2",
						},
					},
				},
				InstancesBeta: []*daisy.InstanceBeta{
					{
						Instance: computeBeta.Instance{
							MachineType: "e2-standard-2",
						},
					},
				},
			},
		},
	}
	return w
}
