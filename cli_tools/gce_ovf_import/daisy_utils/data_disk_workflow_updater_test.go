//  Copyright 2019 Google Inc. All Rights Reserved.
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
//  limitations under the License

package daisyovfutils

import (
	"fmt"
	"testing"

	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	"github.com/stretchr/testify/assert"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/disk"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/domain"
)

func TestAppendDisksToInstance(t *testing.T) {
	wfPath := "../../../daisy_workflows/ovf_import/create_instance.wf.json"
	wf, err := daisy.NewFromFile(wfPath)
	if err != nil {
		t.Fatal(err)
	}

	disks := []domain.Disk{}

	for i := 0; i < 2; i++ {
		diskName := fmt.Sprintf("disk-name%d", i+1)
		disk, err := disk.NewDisk("project", "zone", diskName)
		assert.NoError(t, err)
		disks = append(disks, disk)
	}
	createInstanceStep := wf.Steps["create-instance"].CreateInstances.Instances[0]
	AppendDisksToInstance(createInstanceStep, disks)
	for i, disk := range disks {
		// Offset by one since template includes the bootdisk as the first element in disk lists.
		dataDiskIndex := i + 1
		dataDisk := createInstanceStep.Disks[dataDiskIndex]
		expectedSourceURI := disk.GetURI()
		assert.True(t, dataDisk.AutoDelete)
		assert.Equal(t, expectedSourceURI, dataDisk.Source)
	}
}
