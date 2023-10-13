# Copyright 2021 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM launcher.gcr.io/google/debian11
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -q -y qemu-utils gnupg ca-certificates curl

# Install gcsfuse, installed using instructions from:
# https://cloud.google.com/storage/docs/gcsfuse-install
ENV GCSFUSE_REPO=gcsfuse-bullseye
RUN echo "deb https://packages.cloud.google.com/apt $GCSFUSE_REPO main" > /etc/apt/sources.list.d/gcsfuse.list
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -q -y gcsfuse

COPY linux/gce_ovf_export /gce_ovf_export
COPY daisy_workflows/ /daisy_workflows/
COPY proto/ /proto/

ENTRYPOINT ["/gce_ovf_export"]
