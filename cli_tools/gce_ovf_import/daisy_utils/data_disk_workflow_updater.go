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
	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	"google.golang.org/api/compute/v1"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/domain"
)

// AppendDisksToInstance appends disks to the instance.
func AppendDisksToInstance(instance *daisy.Instance, disks []domain.Disk) {
	for _, disk := range disks {
		instance.Disks = append(instance.Disks,
			&compute.AttachedDisk{
				Source:     disk.GetURI(),
				AutoDelete: true,
			})
	}
}
