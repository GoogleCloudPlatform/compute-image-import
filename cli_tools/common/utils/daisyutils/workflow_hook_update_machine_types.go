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
	"fmt"
	"strings"

	daisy "github.com/GoogleCloudPlatform/compute-daisy"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/logging"
)

// UpdateMachineTypesHook updates the workflow to use the machine series specified as primary.
// If the workflow fails due to the usage quota then it falls back to secondary machine series.
// See cli_tools/common/utils/param/machine_series_detector.go for details.
type UpdateMachineTypesHook struct {
	logger                 logging.Logger
	shouldFallback         bool
	primaryMachineSeries   string
	secondaryMachineSeries string
}

// PreRunHook modifies the workflow to use machine series specified as primary or
// falls back to secondary machine series when needed.
func (f *UpdateMachineTypesHook) PreRunHook(wf *daisy.Workflow) error {
	if f.shouldFallback && f.secondaryMachineSeries != "" {
		f.updateWorkflowMachineSeries(wf, f.secondaryMachineSeries)
		return nil
	}

	f.updateWorkflowMachineSeries(wf, f.primaryMachineSeries)
	return nil
}

// PostRunHook inspects the workflow error to see if it's related to an insufficient primary machine series CPUs quota.
// If so, it requests a retry that will re-run the workflow using the secondary machine series.
func (f *UpdateMachineTypesHook) PostRunHook(err error) (wantRetry bool, wrapped error) {
	if f.secondaryMachineSeries == "" {
		// We cannot fall back if second machine series is not set.
		return false, err
	}

	if f.shouldFallback {
		// A fallback has already occurred; don't request retry.
		return false, err
	} else if err != nil && strings.Contains(err.Error(), fmt.Sprintf("%s_CPUS", strings.ToUpper(f.primaryMachineSeries))) {
		msg := fmt.Sprintf("Workflow failed with an insufficient %s CPUs quota. Requesting retry with %s CPUs. See %s for details.",
			f.primaryMachineSeries,
			f.secondaryMachineSeries,
			"https://cloud.google.com/compute/docs/troubleshooting/troubleshooting-import-export-images")
		f.logger.Debug(msg)
		f.shouldFallback = true
	}
	return f.shouldFallback, err
}

func (f *UpdateMachineTypesHook) updateWorkflowMachineSeries(wf *daisy.Workflow, newSeries string) {
	wf.IterateWorkflowSteps(func(step *daisy.Step) {
		if step.CreateInstances != nil {
			for _, instance := range step.CreateInstances.Instances {
				newMachineType, err := f.updateMachineSeries(instance.MachineType, newSeries)
				if err != nil {
					msg := fmt.Sprintf("Machine type %s was not updated: %s", instance.MachineType, err.Error())
					f.logger.Debug(msg)
					continue
				}

				instance.MachineType = newMachineType
			}

			for _, instance := range step.CreateInstances.InstancesBeta {
				newMachineType, err := f.updateMachineSeries(instance.MachineType, newSeries)
				if err != nil {
					msg := fmt.Sprintf("Machine type %s was not updated: %s", instance.MachineType, err.Error())
					f.logger.Debug(msg)
					continue
				}

				instance.MachineType = newMachineType
			}
		}
	})
}

func (f *UpdateMachineTypesHook) updateMachineSeries(machineType string, newSeries string) (string, error) {
	curSeries, err := f.getMachineSeries(machineType)
	if err != nil {
		return "", err
	}

	return newSeries + machineType[len(curSeries):], nil
}

func (f *UpdateMachineTypesHook) getMachineSeries(machineType string) (string, error) {
	dashIdx := strings.Index(machineType, "-")
	if dashIdx == -1 {
		return "", fmt.Errorf("unknown machine type: %s", machineType)
	}

	return machineType[:dashIdx], nil
}
