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

package param

import (
	daisyCompute "github.com/GoogleCloudPlatform/compute-daisy/compute"
)

// To rebuild the mock for MachineSeriesDetector, run `go generate ./...`
//go:generate go run github.com/golang/mock/mockgen -package mocks -source $GOFILE -destination ../../../mocks/mock_machine_series_detector.go

// MachineSeriesDetector detects which of N2, N1, E2 series are available
// in the execution context and are compatible with the import / export tools.
// N2 https://cloud.google.com/compute/docs/general-purpose-machines#n2_machines
// N1 https://cloud.google.com/compute/docs/general-purpose-machines#n1_machines
// E2 https://cloud.google.com/compute/docs/general-purpose-machines#e2_machines
type MachineSeriesDetector interface {
	// Detects which of N2, N1, E2 series are available in the specified project and zone
	// and are compatible with the import / export tools.
	Detect(project, zone string) ([]string, error)
}

// NewMachineSeriesDetector returns a MachineSeriesDetector implementation that uses the Compute API.
func NewMachineSeriesDetector(client daisyCompute.Client) MachineSeriesDetector {
	return &computeMachineSeriesDetector{client}
}

// computeMachineSeriesDetector uses the Compute API to implement MachineSeriesDetector.
type computeMachineSeriesDetector struct {
	client daisyCompute.Client
}

func (cm *computeMachineSeriesDetector) Detect(project, zone string) ([]string, error) {
	machineTypes, err := cm.getMachineTypes(project, zone)
	if err != nil {
		return []string{}, err
	}

	res := []string{}

	for _, ms := range []string{"n2", "n1", "e2"} {
		if cm.isMachineSeriesCompatible(machineTypes, ms) {
			res = append(res, ms)
		}
	}

	return res, nil
}

func (cm *computeMachineSeriesDetector) getMachineTypes(project, zone string) (map[string]bool, error) {
	machineTypes, err := cm.client.ListMachineTypes(project, zone)
	if err != nil {
		return nil, err
	}

	res := make(map[string]bool)
	for _, mt := range machineTypes {
		res[mt.Name] = true
	}
	return res, nil
}

func (cm *computeMachineSeriesDetector) isMachineSeriesCompatible(machineTypes map[string]bool, machineSeries string) bool {
	// Iterating through the list of all machine types used in the import / export tools.
	for _, mtype := range []string{"-standard-2", "-standard-4", "-standard-8", "-highcpu-4"} {
		if _, ok := machineTypes[machineSeries+mtype]; !ok {
			return false
		}
	}

	return true
}
