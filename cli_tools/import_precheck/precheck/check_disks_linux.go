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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/mount"
	"github.com/GoogleCloudPlatform/osconfig/packages"
)

// DisksCheck performs disk configuration checking:
// - finding the root filesystem partition
// - checking if the device is MBR
// - checking whether the root mount is physically located on a single disk.
//   The check fails, for example, when the root mount is on an LVM
//   logical volume that spans multiple disks.
// - check for GRUB
// - warning for any mount points from partitions from other devices
type DisksCheck struct {
	getMBROverride func(devName string) ([]byte, error)
	inspector      mount.Inspector
}

// NewDisksCheck instantiates a DisksCheck instance.
func NewDisksCheck() Check {
	return &DisksCheck{inspector: mount.NewMountInspector()}
}

// GetName returns the name of the precheck step; this is shown to the user.
func (c *DisksCheck) GetName() string {
	return "Disks Check"
}

// Run executes the precheck step.
func (c *DisksCheck) Run() (r *Report, err error) {
	r = &Report{name: c.GetName()}

	mountInfo, err := getBootDiskMountInfo(r, c.inspector)

	if err != nil {
		return r, nil
	}

	bootDisk := mountInfo.UnderlyingBlockDevices[0]

	// check Partition Table
	outBytes, err := exec.Command("gdisk", "-l", bootDisk).Output()
	if err != nil {
		r.Warn(fmt.Sprintf("Can not run gdisk cmd, %s", err.Error()))
		return r, err
	}
	output := string(outBytes)

	addPartitionTableInfo(r, output)

	// GRUB checking.
	var mbrData []byte
	if c.getMBROverride != nil {
		mbrData, err = c.getMBROverride(bootDisk)
	} else {
		mbrData, err = c.getMBR(bootDisk)
	}

	// get installed pkgs
	ctx := context.Background()
	pkgs, err := packages.GetInstalledPackages(ctx)
	if err != nil {
		return r, fmt.Errorf("GetInstalledPackages error: %s", err)
	}

	addGrubInfo(r, mbrData, pkgs)

	return r, nil
}

func getBootDiskMountInfo(r *Report, inspector mount.Inspector) (mount.InspectionResults, error) {
	mountInfo, err := inspector.Inspect("/")

	if err != nil {
		r.result = Unknown
		r.Warn("Failed to inspect the boot disk. Prior to importing, verify that the boot disk " +
			"contains the root filesystem, and that the root filesystem isn't virtualized over " +
			"multiple disks (using LVM, for example).")
		return mountInfo, err
	}

	r.Info(fmt.Sprintf("root filesystem mounted on %s", mountInfo.BlockDevicePath))

	if len(mountInfo.UnderlyingBlockDevices) > 1 {
		format := "root filesystem spans multiple block devices (%s). Typically this occurs when an LVM logical " +
			"volume spans multiple block devices. Image import only supports single block device."
		r.Fatal(fmt.Sprintf(format, strings.Join(mountInfo.UnderlyingBlockDevices, ", ")))
		return mountInfo, fmt.Errorf("Root filesystem spans multiple block devices")
	}

	r.Info(fmt.Sprintf("boot disk detected as %s", mountInfo.UnderlyingBlockDevices[0]))

	return mountInfo, err
}

func addGrubInfo(r *Report, mbrData []byte, pkgs *packages.Packages) {
	if bytes.Contains(mbrData, []byte("GRUB")) || hasGrub2(pkgs.Rpm) || hasGrub2(pkgs.Deb) {
		r.Info("GRUB detected")
		return
	}

	r.Warn("GRUB not detected")
}

func hasGrub2(pkgs []packages.PkgInfo) bool {
	for _, pkg := range pkgs {
		if strings.Index(pkg.Name, "grub2-") != -1 {
			return true
		}
	}
	return false
}

func addPartitionTableInfo(r *Report, gdiskCommandResult string) {
	if regexp.MustCompile("(.*)GPT:[\\s]*present(.*)").MatchString(gdiskCommandResult) {
		if regexp.MustCompile("(.*)MBR:[\\s]*protective(.*)").MatchString(gdiskCommandResult) {
			r.Info("boot disk has valid GPT with protective MBR; using GPT.")
		} else if regexp.MustCompile("(.*)MBR:[\\s]*hybrid(.*)").MatchString(gdiskCommandResult) {
			r.Info("boot disk has valid GPT with hybrid MBR; using GPT.")
		} else if regexp.MustCompile("(.*)MBR:[\\s]*not present(.*)").MatchString(gdiskCommandResult) {
			r.Info("boot disk has GPT only.")
		} else {
			r.Warn("Unkown boot disk partition table.")
		}
	} else if regexp.MustCompile("(.*)MBR:[\\s]*MBR only(.*)").MatchString(gdiskCommandResult) {
		r.Info("boot disk has MBR only.")
	} else {
		r.Warn("Unkown boot disk partition table.")
	}
}

func (c *DisksCheck) getMBR(devPath string) ([]byte, error) {
	f, err := os.Open(devPath)
	if err != nil {
		return nil, err
	}
	data := make([]byte, mbrSize)
	_, err = f.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", devPath, err)
	}
	return data, nil
}
