#!/bin/bash
# Copyright 2022 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Runs validation tests for Debian 11 worker image. Script is executed in
# workflow debian_11_worker.wf.json

echo "BuildStatus: Test if apt packages have been installed"
APT_PACKAGES="
debootstrap
dosfstools
kpartx
libguestfs-tools
parted
python3-guestfs
python3-netaddr
python3-pip
rsync
tinyproxy
qemu-utils
"
dpkg -l ${APT_PACKAGES}
if [[ $? -ne 0 ]]; then
  echo "BuildFailed: Not all packages were found."
  exit 1
fi

echo "BuildStatus: Test if all python3 pip libraries have been installed"
PIP3_PACKAGES="google-api-python-client google-cloud-storage"

for pip_package in $PIP3_PACKAGES
do
  pip3 list | grep -q $pip_package
  if [[ $? -ne 0 ]]; then
    echo "BuildFailed: Not all python3 pip libraries found."
    exit 1
  fi
done

echo "BuildStatus: Check if gce_export is available"
if [[ ! -f "/usr/bin/gce_export" ]]; then
  echo "BuildFailed: Could not find gce_export."
  exit 1
fi

echo "BuildStatus: Check if gce_export is executable"
if [[ ! -x "/usr/bin/gce_export" ]]; then
  echo "BuildFailed: gce_export is not executable."
  exit 1 
fi

echo "BuildSuccess: Worker tests succeeded."
