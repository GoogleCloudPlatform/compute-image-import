//  Copyright 2017 Google Inc. All Rights Reserved.
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

package precheck

import (
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/mount"
	"github.com/GoogleCloudPlatform/osconfig/packages"

	"github.com/stretchr/testify/assert"
)

type disksCheckTest struct {
	name           string
	mountInfo      mount.InspectionResults
	inspectError   error
	byteTrailer    []byte
	expectAllLogs  []string
	expectedStatus Result
}

type partitionTableInfoTest struct {
	name               string
	gdiskCommandResult string
	expectInfo         string
}

func TestDisksCheck_Inspector(t *testing.T) {
	for _, tc := range []disksCheckTest{
		{
			name: "pass if boot device is non virtual",
			mountInfo: mount.InspectionResults{
				BlockDevicePath:        "/dev/sda1",
				BlockDeviceIsVirtual:   false,
				UnderlyingBlockDevices: []string{"/dev/sda"},
			},
			expectAllLogs: []string{
				"INFO: root filesystem mounted on /dev/sda1",
			},
			expectedStatus: Passed,
		}, {
			name: "pass if boot device is virtual with one underlying device",
			mountInfo: mount.InspectionResults{
				BlockDevicePath:        "/dev/mapper/vg-lv",
				BlockDeviceIsVirtual:   true,
				UnderlyingBlockDevices: []string{"/dev/sda"},
			},
			expectAllLogs: []string{
				"INFO: root filesystem mounted on /dev/mapper/vg-lv",
			},
			expectedStatus: Passed,
		}, {
			name: "fail if boot device is virtual with multiple underlying devices",
			mountInfo: mount.InspectionResults{
				BlockDevicePath:        "/dev/mapper/vg-lv",
				BlockDeviceIsVirtual:   true,
				UnderlyingBlockDevices: []string{"/dev/sda", "/dev/sdb"},
			},
			expectAllLogs: []string{
				"FATAL: root filesystem spans multiple block devices (/dev/sda, /dev/sdb). Typically this occurs when an LVM " +
					"logical volume spans multiple block devices. Image import only supports single block device.",
			},
			expectedStatus: Failed,
		}, {
			name:         "fail if inspect fails",
			inspectError: errors.New("failed to find root device"),
			expectAllLogs: []string{
				"WARN: Failed to inspect the boot disk. Prior to importing, verify that the boot disk " +
					"contains the root filesystem, and that the root filesystem isn't virtualized " +
					"over multiple disks (using LVM, for example).",
			},
			expectedStatus: Unknown,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var report Report
			addBootDiskMountInfo(&report, tc.mountInfo, tc.inspectError != nil)
			for _, log := range tc.expectAllLogs {
				assert.Contains(t, report.logs, log)
			}
		})
	}
}

func Test_addGrubInfo_GrubDetected_bytes(t *testing.T) {
	pkgs := packages.Packages{}
	bytes := []byte{'G', 'R', 'U', 'B', 0x55, 0xAA}
	var report Report
	addGrubInfo(&report, bytes, &pkgs)
	assert.Contains(t, report.logs, "INFO: GRUB detected")
}

func Test_addGrubInfo_GrubDetected_Pkgs(t *testing.T) {
	rpmPkgs := []packages.PkgInfo{{Name: "grub2-common", Arch: "all", Version: "1:2.02-0.87.el7_9.7"}}
	pkgs := packages.Packages{Rpm: rpmPkgs}
	bytes := []byte{0x55, 0xAA}
	var report Report
	addGrubInfo(&report, bytes, &pkgs)
	assert.Contains(t, report.logs, "INFO: GRUB detected")
}

func Test_addGrubInfo_GrubNotDetected(t *testing.T) {
	rpmPkgs := []packages.PkgInfo{{Name: "chkconfig", Arch: "x86_64", Version: "1.7.6-1.el7"}}
	pkgs := packages.Packages{Rpm: rpmPkgs}
	bytes := []byte{0x55, 0xAA}
	var report Report
	addGrubInfo(&report, bytes, &pkgs)
	assert.Contains(t, report.logs, "WARN: GRUB not detected")
}

func Test_addPartitionTableInfo(t *testing.T) {

	for _, tc := range []partitionTableInfoTest{
		{
			name:               "GPT With Protective MBR",
			gdiskCommandResult: "Partition table scan:\n  MBR: protective\n  BSD: not present\n  APM: not present\n  GPT: present \n",
			expectInfo:         "INFO: boot disk has valid GPT with protective MBR; using GPT.",
		}, {
			name:               "MBR Only",
			gdiskCommandResult: "Partition table scan:\n  MBR: MBR only\n  BSD: not present\n  APM: not present\n  GPT: not present",
			expectInfo:         "INFO: boot disk has MBR only.",
		}, {
			name:               "GPT Only",
			gdiskCommandResult: "Partition table scan:\n  MBR: not present\n  BSD: not present\n  APM: not present\n  GPT: present\n",
			expectInfo:         "INFO: boot disk has GPT only.",
		}, {
			name:               "GPT With Hybrid MBR",
			gdiskCommandResult: "Partition table scan:\n  MBR: hybrid\n  BSD: not present\n  APM: not present\n  GPT: present\n",
			expectInfo:         "INFO: boot disk has valid GPT with hybrid MBR; using GPT.",
		}, {
			name:               "Unkown boot disk partition table",
			gdiskCommandResult: "Partition table scan:\n  MBR: not present\n  BSD: not present\n  APM: not present\n  GPT: not present\n",
			expectInfo:         "WARN: Unkown boot disk partition table.",
		}, {
			name:               "Unkown boot disk partition table2",
			gdiskCommandResult: "Partition table scan:\n  MBR: corrupted-text\n  BSD: not present\n  APM: not present\n  GPT: present\n",
			expectInfo:         "WARN: Unkown boot disk partition table.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var report Report
			addPartitionTableInfo(&report, tc.gdiskCommandResult)
			assert.Contains(t, report.logs, tc.expectInfo)
		})
	}
}
