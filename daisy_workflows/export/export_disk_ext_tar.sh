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

set +x
serialOutputPrefixedKeyValue "GCEExport" "source-size-gb" "${SIZE_OUTPUT_GB}"
set -x

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

echo "GCEExport: Exporting disk of size ${SIZE_OUTPUT_GB}GB and format ${FORMAT}."

PART_FILES=()
function cleanup() {
  if [[ "${#PART_FILES[@]}" -eq 0 ]]; then
    return
  fi
  echo "GCEExport: Cleaning up part files: ${PART_FILES[*]}"
  if docker run --rm "${GCLOUD_CLI_IMAGE}" gcloud storage rm "${PART_FILES[@]}"; then
    PART_FILES=()
  else
    echo "ExportWarning: Failed to remove files: ${PART_FILES[*]}"
  fi
}
trap cleanup EXIT

cd "/var/gs"

# 1. Create empty disk.raw file of right size. This is just used to create the correct tar header.
if ! truncate -s "${SIZE_BYTES}" disk.raw; then
  echo "ExportFailed: Failed to create empty disk.raw file."
  exit
fi

# 2. Steal the 512-byte header, forcing GNU format to ensure even large tars have a single 512-byte header.
if ! tar --create --format=gnu --owner=0 --group=0 --mode=0600 --file - disk.raw | head -c 512 > header.bin; then
  echo "ExportFailed: Failed to create tar header."
  exit
fi
rm disk.raw

# 3. Calculate alignment padding + 1024 bytes for the Tar EOF marker
REMAINDER=$(( SIZE_BYTES % 512 ))
PAD_BYTES=$(( REMAINDER == 0 ? 0 : 512 - REMAINDER ))
TOTAL_TRAILER_BYTES=$(( PAD_BYTES + 1024 ))

# 4. Split into chunks and stream them separately for performance
# GZIPs can be concatated together into a single valid GZIP file, and doing so has minimal impact on compression.

CPU_COUNT=$(nproc)
if [[ $CPU_COUNT -gt 8 ]]; then
  CPU_COUNT=8 # limit to reduce risk of hitting OOMs.
fi

for (( i=0; i<CPU_COUNT; i++ )); do
  PART_FILES+=("${GS_PATH}.part${i}")
done

CHUNK_SIZE=$(( SIZE_BYTES / CPU_COUNT ))

echo "GCEExport: TOTAL_TRAILER_BYTES: ${TOTAL_TRAILER_BYTES}, CPU_COUNT: ${CPU_COUNT}, CHUNK_SIZE: ${CHUNK_SIZE}"

PIDS=()
for (( i=0; i<CPU_COUNT; i++ )); do
  (
    set -eo pipefail
    SKIP_BYTES=$(( i * CHUNK_SIZE ))
    COUNT_BYTES=$CHUNK_SIZE
    if [[ $i -eq $(( CPU_COUNT - 1 )) ]]; then
      COUNT_BYTES=$(( SIZE_BYTES - SKIP_BYTES ))
    fi

    echo "GCEExport chunk ${i}: SKIP_BYTES: ${SKIP_BYTES}, COUNT_BYTES: ${COUNT_BYTES}"

    (
      # Stream header for the first chunk.
      if [[ $i -eq 0 ]]; then
        cat header.bin
      fi

      # Stream the chunk.
      dd if="${SOURCE_DEVICE}" bs=4M skip="${SKIP_BYTES}" count="${COUNT_BYTES}" iflag=skip_bytes,count_bytes

      # Stream the trailer for the last chunk.
      if [[ $i -eq $(( CPU_COUNT - 1 )) ]]; then
        head -c "${TOTAL_TRAILER_BYTES}" /dev/zero
      fi
    ) | gzip -3 -c | docker run -i --rm "${GCLOUD_CLI_IMAGE}" gcloud storage cp --quiet - "${GS_PATH}.part${i}"
  ) &
  PIDS+=($!)
done

(
  while sleep 100; do
    for (( i=0; i<CPU_COUNT; i++ )); do
      SKIP_BYTES=$(( i * CHUNK_SIZE ))
      if DD_PID=$(pgrep -f "dd if=${SOURCE_DEVICE}.*skip=${SKIP_BYTES} "); then
        echo "--- Progress for chunk ${i} ---"
        # Send SIGUSR1 to the dd process to print progress to stderr.
        kill -USR1 "$DD_PID"
        # Sleep so that progress from different chunks don't overlap.
        sleep 5
      fi
    done
  done
) &
MONITOR_PID=$!

FAIL=0
for pid in "${PIDS[@]}"; do
  wait "$pid" || FAIL=1
done

kill "$MONITOR_PID"

if [[ $FAIL -ne 0 ]]; then
  echo "ExportFailed: Failed to export disk source to GCS [Privacy-> ${GS_PATH} <-Privacy] via native streaming tar.gz"
  exit
fi

if ! docker run --rm "${GCLOUD_CLI_IMAGE}" gcloud storage objects compose "${PART_FILES[@]}" "${GS_PATH}"; then
  echo "ExportFailed: Failed to compose parts into ${GS_PATH}"
  exit
fi

cleanup

rm header.bin

# Exported image size info.
TARGET_SIZE_BYTES=$(docker run --rm "${GCLOUD_CLI_IMAGE}" gcloud storage du -s "${GS_PATH}" | awk '{print $1}')
TARGET_SIZE_GB=$(awk "BEGIN {print int(((${TARGET_SIZE_BYTES}-1)/${BYTES_1GB}) + 1)}")
set +x
serialOutputPrefixedKeyValue "GCEExport" "target-size-gb" "${TARGET_SIZE_GB}"
set -x

# TODO(b/460360483): Change the success/failure signal as sometimes the last
# lines are not printed. Use another signal - https://github.com/GoogleCloudPlatform/compute-daisy/blob/master/step_wait_for_instances_signal.go#L81.
echo "export success"
sleep 5
echo "export success"
sleep 5
echo "export success"

sync
