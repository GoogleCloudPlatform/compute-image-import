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
	"context"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/storage"
	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/param"
	"google.golang.org/api/option"
)

// ApplyEnvToWorkflow is a WorkflowHook that applies user-customizable values
// to the top-level parent workflow.
type ApplyEnvToWorkflow struct {
	env EnvironmentSettings
}

// PreRunHook updates properties on wf that correspond to user-specified values
// such as project, zone, and scratch bucket path.
func (t *ApplyEnvToWorkflow) PreRunHook(wf *daisy.Workflow) error {
	set(t.env.Project, &wf.Project)
	set(t.env.Zone, &wf.Zone)
	set(t.env.GCSPath, &wf.GCSPath)
	set(t.env.OAuth, &wf.OAuthPath)
	set(t.env.Timeout, &wf.DefaultTimeout)

	err := updateWorkflowClientsIfneeded(t.env, wf)

	return err
}

func set(src string, dst *string) {
	if src != "" {
		*dst = src
	}
}

func updateWorkflowClientsIfneeded(env EnvironmentSettings, wf *daisy.Workflow) error {

	// Create new context here as daisy clients shouldn't die if the main context is cancelled.
	// It will need to cleaned up resources before terminating the clients.
	ctx := context.Background()

	if env.EndpointsOverride.Compute != "" {
		daisyComputeClient, err := param.CreateComputeClient(&ctx, env.OAuth, env.EndpointsOverride.Compute)
		if err != nil {
			return err
		}
		wf.ComputeClient = daisyComputeClient
	}

	if env.EndpointsOverride.Storage != "" {
		storageOptions := []option.ClientOption{option.WithEndpoint(env.EndpointsOverride.Storage)}
		if env.OAuth != "" {
			storageOptions = append(storageOptions, option.WithCredentialsFile(env.OAuth))
		}

		storageClient, err := storage.NewClient(ctx, storageOptions...)
		if err != nil {
			return err
		}

		wf.StorageClient = storageClient
	}

	if env.EndpointsOverride.CloudLogging != "" {
		cloudLoggingOptions := []option.ClientOption{option.WithEndpoint(env.EndpointsOverride.CloudLogging)}
		if env.OAuth != "" {
			cloudLoggingOptions = append(cloudLoggingOptions, option.WithCredentialsFile(env.OAuth))
		}

		cloudLoggingClient, err := logging.NewClient(ctx, wf.Project, cloudLoggingOptions...)
		if err != nil {
			return err
		}

		wf.CloudLoggingClient = cloudLoggingClient
	}

	return nil
}
