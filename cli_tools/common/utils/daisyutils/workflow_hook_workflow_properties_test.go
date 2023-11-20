//  Copyright 2021 Google Inc. All Rights Reserved.
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

package daisyutils

import (
	"testing"

	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	"github.com/stretchr/testify/assert"
)

func Test_ApplyEnvToWorkflow(t *testing.T) {
	env := EnvironmentSettings{
		Project: "lucky-lemur",
		Zone:    "us-west1-c",
		GCSPath: "new-path",
		OAuth:   "new-oauth",
		Timeout: "new-timeout",
	}
	original := &daisy.Workflow{
		Project:        "original-project",
		Zone:           "original-zone",
		GCSPath:        "original-path",
		OAuthPath:      "original-oauth",
		DefaultTimeout: "original-timeout",
	}
	assert.NoError(t, (&ApplyEnvToWorkflow{env}).PreRunHook(original))
	expected := &daisy.Workflow{
		Project:        "lucky-lemur",
		Zone:           "us-west1-c",
		GCSPath:        "new-path",
		OAuthPath:      "new-oauth",
		DefaultTimeout: "new-timeout",
	}
	assert.Equal(t, original, expected)
}

func Test_updateWorkflowClientsIfneeded_OverrideClients(t *testing.T) {
	env := EnvironmentSettings{
		EndpointsOverride: EndpointsOverride{Compute: "https://compute.googleapis.com/compute/v1/",
			Storage: "https://storage.googleapis.com/storage/v1/", CloudLogging: "https://logging.googleapis.com/logging/v1/"},
	}

	wf := &daisy.Workflow{
		Project:        "project",
		Zone:           "zone",
		GCSPath:        "path",
		OAuthPath:      "oauth",
		DefaultTimeout: "timeout",
	}

	err := updateWorkflowClientsIfneeded(env, wf)
	assert.NoError(t, err)
	assert.NotNil(t, wf.ComputeClient)
	assert.NotNil(t, wf.StorageClient)
	assert.NotNil(t, wf.CloudLoggingClient)
}

func Test_updateWorkflowClientsIfneeded_dontOverrideClients(t *testing.T) {
	env := EnvironmentSettings{}

	wf := &daisy.Workflow{
		Project:        "project",
		Zone:           "zone",
		GCSPath:        "path",
		OAuthPath:      "oauth",
		DefaultTimeout: "timeout",
	}

	err := updateWorkflowClientsIfneeded(env, wf)
	assert.NoError(t, err)
	assert.Nil(t, wf.ComputeClient)
	assert.Nil(t, wf.StorageClient)
	assert.Nil(t, wf.CloudLoggingClient)
}
