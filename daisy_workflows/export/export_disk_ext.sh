#!/bin/bash
# Copyright 2017 Google Inc. All Rights Reserved.
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
set -x

function serialOutputPrefixedKeyValue() {
  stdbuf -oL echo "$1: <serial-output key:'$2' value:'$3'>"
}

GCLOUD_CLI_IMAGE="gcr.io/google.com/cloudsdktool/google-cloud-cli:545.0.0-slim"

function disk_resizing_monitor() {
  local max_size=$1
  local buffer_min_size=10
  local buffer_size=25
  local interval=10

  echo "GCEExport: Max disk size ${max_size}GB, min buffer size ${buffer_size}GB, starting monitoring available disk buffer every ${interval}s..."
  while sleep ${interval}; do
    # Check whether available buffer space is lower than threshold.
    local available_buffer
    available_buffer=$(df -BG "${BUFFER_DEVICE}" --output=avail | sed -n 2p)
    available_buffer=${available_buffer%?}
    if [[ ${available_buffer} -ge ${buffer_min_size} ]]; then
      continue
    fi

    # Decide the new size of the device.
    local current_device_size_bytes
    current_device_size_bytes=$(lsblk "${BUFFER_DEVICE}" --output=size -b | sed -n 2p)
    local current_device_size_gb
    current_device_size_gb=$(awk "BEGIN {print int(((${current_device_size_bytes}-1)/${BYTES_1GB}) + 1)}")
    local next_size_gb
    next_size_gb=$(awk "BEGIN {print int(${current_device_size_gb} + ${buffer_size})}")

    echo "GCEExport: Resizing buffer disk from ${current_device_size_gb}GB to ${next_size_gb}GB..."
    # This command assumes that the docker configuration was done already in the
    # export_disk_ext.sh script.
    if ! out=$(docker run --rm "${GCLOUD_CLI_IMAGE}" gcloud compute disks resize "${BUFFER_DISK}" --size="${next_size_gb}GB" --quiet --zone "${ZONE}" 2>&1); then
      echo "ExportFailed: Failed to resize buffer disk. [Privacy-> Error: ${out} <-Privacy]"
      continue
    fi
    echo "GCEExport: ${out}"
    if ! out=$(sudo resize2fs "${BUFFER_DEVICE}" 2>&1); then
      echo "ExportFailed: Failed to resize partition of buffer disk. [Privacy-> Error: ${out} <-Privacy]"
      continue
    fi
    echo "${out}"

    # If current file system has reached or exceeded max size, then stop resizing.
    # We need to know the size of the available file system other than the size of
    # the partition, so "df" is used instead of "lsblk" here.
    local current_filesystem_size
    current_filesystem_size=$(df -BG "${BUFFER_DEVICE}" --output=size | sed -n 2p)
    current_filesystem_size=${current_filesystem_size%?}
    if [[ ${current_filesystem_size} -ge ${max_size} ]]; then
      echo "Buffer disk reaches max size."
      continue
    fi
  done
}

# Verify VM has access to Google APIs
curl --silent --fail "https://www.googleapis.com/discovery/v1/apis" &> /dev/null;
if [[ $? -ne 0 ]]; then
  echo "ExportFailed: Cannot access Google APIs. Ensure that VPC settings allow VMs to access Google APIs either via external IP or Private Google Access. More info at: https://cloud.google.com/vpc/docs/configure-private-google-access"
  exit
fi

BYTES_1GB=1073741824
METADATA_URL="http://169.254.169.254/computeMetadata/v1/instance"
ATTRIBUTES_URL="${METADATA_URL}/attributes"
GS_PATH=$(curl -f -H Metadata-Flavor:Google ${ATTRIBUTES_URL}/gcs-path)
FORMAT=$(curl -f -H Metadata-Flavor:Google ${ATTRIBUTES_URL}/format)

# Strip gs:// prefix.
IMAGE_OUTPUT_PATH=${GS_PATH##*//}
# Create dir for output.
OUTS_PATH=${IMAGE_OUTPUT_PATH%/*}
mkdir -p "/var/gs/${OUTS_PATH}"

# Prepare disk size info.
# 1. Disk image size info.
SOURCE_DEVICE=$(readlink -f /dev/disk/by-id/google-disk-image-export-ext)
SIZE_BYTES=$(lsblk "${SOURCE_DEVICE}" --output=size -b | sed -n 2p)
# 2. Round up to the next GB.
SIZE_OUTPUT_GB=$(awk "BEGIN {print int(((${SIZE_BYTES}-1)/${BYTES_1GB}) + 1)}")
# 3. Add 5GB of additional space to max size to prevent the corner case that output
# file is slightly larger than source disk.
MAX_BUFFER_DISK_SIZE_GB=$(awk "BEGIN {print int(${SIZE_OUTPUT_GB} + 5)}")

set +x
serialOutputPrefixedKeyValue "GCEExport" "source-size-gb" "${SIZE_OUTPUT_GB}"
set -x

# Prepare buffer disk.
echo "GCEExport: Initializing buffer disk for qemu-img output..."
BUFFER_DEVICE=$(readlink -f /dev/disk/by-id/google-disk-export-disk-buffer*)
mkfs.ext4 "${BUFFER_DEVICE}"
mount "${BUFFER_DEVICE}" "/var/gs/${OUTS_PATH}"
if [[ $? -ne 0 ]]; then
  echo "ExportFailed: Failed to prepare buffer disk by mkfs + mount."
fi

# Prepare parameters for resizing
ZONE=$(curl "${METADATA_URL}/zone" -H "Metadata-Flavor: Google"| cut -d'/' -f4)
BUFFER_DISK=$(curl "${ATTRIBUTES_URL}/buffer-disk" -H "Metadata-Flavor: Google")

# Configure Docker to use the pre-installed docker-credential-gcr helper.
# This command will write the configuration to $DOCKER_CONFIG/config.json.
# The helper uses the VM's service account credentials from the metadata server.
# Log in to GCR using the fetched token.
# We need to set the HOME environment variable to /var/gs as COS the FS is read-only.
export HOME=/var/gs
if ! out=$(/usr/bin/docker-credential-gcr configure-docker); then
  echo "ExportFailed: Failed to configure docker. [Privacy-> Error: ${out} <-Privacy]"
  exit
fi

echo "GCEExport: Pulling docker image ${GCLOUD_CLI_IMAGE}..."
if ! out=$(docker pull "${GCLOUD_CLI_IMAGE}" 2>&1); then
  echo "ExportFailed: Failed to pull docker image [Privacy-> ${GCLOUD_CLI_IMAGE} <-Privacy]. Error: [Privacy-> ${out} <-Privacy]"
  exit
fi
echo "${out}"

echo "GCEExport: Launching disk size monitor in background..."
disk_resizing_monitor "${MAX_BUFFER_DISK_SIZE_GB}" &

echo "GCEExport: Exporting disk of size ${SIZE_OUTPUT_GB}GB and format ${FORMAT}."

QEMU_IMG_DOCKER_IMAGE=$(curl -f -H Metadata-Flavor:Google ${ATTRIBUTES_URL}/qemu-img-docker-image)
echo "GCEExport: Pulling docker image ${QEMU_IMG_DOCKER_IMAGE}..."
if ! out=$(docker pull "${QEMU_IMG_DOCKER_IMAGE}" 2>&1); then
  echo "ExportFailed: Failed to pull docker image [Privacy-> ${QEMU_IMG_DOCKER_IMAGE} <-Privacy]. Error: [Privacy-> ${out} <-Privacy]"
  exit
fi
echo "${out}"

echo "GCEExport: Running qemu-img convert..."
docker run --rm -v /tmp:/t -e HOME=/root -v /var/gs:/var/gs --device="${SOURCE_DEVICE}":"${SOURCE_DEVICE}" --privileged "${QEMU_IMG_DOCKER_IMAGE}" /qemu-img convert "${SOURCE_DEVICE}" "/var/gs/${IMAGE_OUTPUT_PATH}" -p -O "${FORMAT}" 2> >(tee /var/gs/qemu_err.txt >&2)
if [[ $? -ne 0 ]]; then
  echo "ExportFailed: Failed to export disk source to GCS [Privacy-> ${GS_PATH} <-Privacy] due to qemu-img error: [Privacy-> $(</var/gs/qemu_err.txt) <-Privacy]"
  exit
fi

# Exported image size info.
TARGET_SIZE_BYTES=$(du -b "/var/gs/${IMAGE_OUTPUT_PATH}" | awk '{print $1}')
TARGET_SIZE_GB=$(awk "BEGIN {print int(((${TARGET_SIZE_BYTES}-1)/${BYTES_1GB}) + 1)}")
set +x
serialOutputPrefixedKeyValue "GCEExport" "target-size-gb" "${TARGET_SIZE_GB}"
set -x

echo "GCEExport: Copying output image to target GCS path..."
docker run --rm -v /var/gs:/var/gs "${GCLOUD_CLI_IMAGE}" gcloud storage cp "/var/gs/${IMAGE_OUTPUT_PATH}" "${GS_PATH}" 2> >(tee /var/gs/gcloud_err.txt >&2)
if [[ $? -ne 0 ]] ; then
  echo "ExportFailed: Failed to copy output image to GCS [Privacy-> ${GS_PATH}, error: $(</var/gs/gcloud_err.txt) <-Privacy]"
  exit
fi

# TODO(b/460360483): Change the success/failure signal as sometimes the last
# lines are not printed. Use another signal - https://github.com/GoogleCloudPlatform/compute-daisy/blob/master/step_wait_for_instances_signal.go#L81.
echo "export success"
sleep 5
echo "export success"
sleep 5
echo "export success"

sync
