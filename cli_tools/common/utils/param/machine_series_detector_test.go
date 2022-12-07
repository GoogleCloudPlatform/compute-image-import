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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/mocks"
)

func Test_RetrievingSupportedMachineSeries(t *testing.T) {
	project := "a-project"
	zone := "a-zone"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockComputeClient := mocks.NewMockClient(mockCtrl)
	mockComputeClient.EXPECT().ListMachineTypes(project, zone).Return([]*compute.MachineType{
		// n1 doesn't have the machine type n1-highcpu-4 in this zone and therefore is skipped
		{Name: "n1-standard-2"},
		{Name: "n1-standard-4"},
		{Name: "n1-standard-8"},

		{Name: "e2-standard-2"},
		{Name: "e2-standard-4"},
		{Name: "e2-standard-8"},
		{Name: "e2-highcpu-4"},

		// n2d is not in the set of (n2, n1, e2)
		{Name: "n2d-standard-2"},
		{Name: "n2d-standard-4"},
		{Name: "n2d-standard-8"},
		{Name: "n2d-highcpu-4"},

		{Name: "n2-standard-2"},
		{Name: "n2-standard-4"},
		{Name: "n2-standard-8"},
		{Name: "n2-highcpu-4"},
	}, nil)

	machineSeriesDetector := NewMachineSeriesDetector(mockComputeClient)

	actual, err := machineSeriesDetector.Detect(project, zone)
	assert.NoError(t, err)

	assert.Equal(t, []string{"n2", "e2"}, actual)
}
