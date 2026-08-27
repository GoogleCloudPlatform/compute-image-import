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
//  limitations under the License.

// Package exporter defines GCE VM image exporter
package exporter

import (
	"context"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	daisy "github.com/GoogleCloudPlatform/compute-daisy"
	daisyCompute "github.com/GoogleCloudPlatform/compute-daisy/compute"
    computeV1 "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/compute"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/daisyutils"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/logging"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/param"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/path"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/storage"
	stringutils "github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/string"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/validation"
	"github.com/GoogleCloudPlatform/compute-image-import/proto/go/pb"
)

// Make file paths mutable
var (
	WorkflowDir       = "daisy_workflows/export/"
	ExportWorkflow    = "image_export_ext.wf.json"
	ExportTarWorkflow = "image_export_ext_tar.wf.json"
)

// Parameter key shared with external packages
const (
	ClientIDFlagKey           = "client_id"
	DestinationURIFlagKey     = "destination_uri"
	SourceImageFlagKey        = "source_image"
	SourceDiskSnapshotFlagKey = "source_disk_snapshot"

	targetSizeGBKey = "target-size-gb"
	sourceSizeGBKey = "source-size-gb"
)

// ImageExportRequest includes the parameters required to perform an image export.
type ImageExportRequest struct {
	ClientID                    string
	DestinationURI              string
	SourceImage                 string
	SourceDiskSnapshot          string
	Format                      string
	Project                     string
	Network                     string
	Subnet                      string
	Zone                        string
	Timeout                     string
	ScratchBucketGcsPath        string
	Oauth                       string
	ComputeEndpoint             string
	ComputeServiceAccount       string
	GcsLogsDisabled             bool
	CloudLogsDisabled           bool
	StdoutLogsDisabled          bool
	Labels                      string
	CurrentExecutablePath       string
	NestedVirtualizationEnabled bool
	WorkerMachineSeries         []string
	QemuImgDockerImage          string
	ExternalIP                  string
}

func validateExternalIP(externalIP string) error {
	if externalIP == "" {
		return nil
	}
	if strings.EqualFold(externalIP, "none") || strings.EqualFold(externalIP, "ephemeral") {
		return nil
	}
	if net.ParseIP(externalIP) != nil {
		return nil
	}
	return daisy.Errf("External IP must be 'ephemeral', 'none', or a valid IP address")
}

func validateAndParseFlags(destinationURI string, sourceImage string, sourceDiskSnapshot string, labels string, externalIP string) (map[string]string, error) {
	if err := validation.ValidateStringFlagNotEmpty(destinationURI, DestinationURIFlagKey); err != nil {
		return nil, err
	}
	if err := validation.ValidateExactlyOneOfStringFlagNotEmpty(map[string]string{
		SourceImageFlagKey:        sourceImage,
		SourceDiskSnapshotFlagKey: sourceDiskSnapshot,
	}); err != nil {
		return nil, err
	}

	if err := validateExternalIP(externalIP); err != nil {
		return nil, err
	}

	if labels != "" {
		userLabels, err := param.ParseKeyValues(labels)
		if err != nil {
			return nil, err
		}
		return userLabels, nil
	}
	return nil, nil
}

func getWorkflowPath(format string, currentExecutablePath string) string {
	workflow := ExportWorkflow
	if format == "tar.gz" {
		workflow = ExportTarWorkflow
	}
	if basePath := os.Getenv("WORKFLOW_BASE_PATH"); basePath != "" {
		return filepath.Join(basePath, workflow)
	}

	return path.ToWorkingDir(WorkflowDir+workflow, currentExecutablePath)
}

func buildDaisyVars(destinationURI string, sourceImage string, sourceDiskSnapshot string, imageDiskSizeGb int64, format string, network string,
	subnet string, region string, computeServiceAccount string, qemuImgDockerImage string) map[string]string {

	destinationURI = strings.TrimSpace(destinationURI)
	sourceImage = strings.TrimSpace(sourceImage)
	sourceDiskSnapshot = strings.TrimSpace(sourceDiskSnapshot)
	format = strings.TrimSpace(format)
	network = strings.TrimSpace(network)
	subnet = strings.TrimSpace(subnet)
	region = strings.TrimSpace(region)
	computeServiceAccount = strings.TrimSpace(computeServiceAccount)
	qemuImgDockerImage = strings.TrimSpace(qemuImgDockerImage)

	varMap := map[string]string{}

	varMap["destination"] = destinationURI

	if sourceImage != "" {
		varMap["source_image"] = param.GetGlobalResourcePath(
			"images", sourceImage)
	}

	if sourceDiskSnapshot != "" {
		varMap["source_disk_snapshot"] = param.GetGlobalResourcePath("snapshots", sourceDiskSnapshot)
	}

	if imageDiskSizeGb > 0 && format != "tar.gz" {
		// We want the size to allow for a 2 step (dd and tar) packaging of the image.
		// In addition, we want to allow for some overhead (5%) in case of completely random data.
		bufferDiskSizeGb := int64(math.Ceil(float64(imageDiskSizeGb) * 2.05))
		varMap["export_instance_disk_size"] = strconv.FormatInt(bufferDiskSizeGb, 10)
	}

	if format != "" {
		varMap["format"] = format
	}
	if subnet != "" {
		varMap["export_subnet"] = param.GetRegionalResourcePath(
			region, "subnetworks", subnet)

		// When subnet is set, we need to grant a value to network to avoid fallback to default
		if network == "" {
			varMap["export_network"] = ""
		}
	}
	if network != "" {
		varMap["export_network"] = param.GetGlobalResourcePath(
			"networks", network)
	}
	if computeServiceAccount != "" {
		varMap["compute_service_account"] = computeServiceAccount
	}
	if qemuImgDockerImage != "" && format != "tar.gz" {
		varMap["qemu_img_docker_image"] = qemuImgDockerImage
	}
	return varMap
}

// Run runs export workflow.
func Run(logger logging.Logger, args *ImageExportRequest) error {

	userLabels, err := validateAndParseFlags(args.DestinationURI, args.SourceImage, args.SourceDiskSnapshot, args.Labels, args.ExternalIP)
	if err != nil {
		return err
	}

	ctx := context.Background()
	metadataGCE := &compute.MetadataGCE{}
	storageClient, err := storage.NewStorageClient(
		ctx, logger, option.WithCredentialsFile(args.Oauth))
	if err != nil {
		return err
	}
	defer storageClient.Close()

	scratchBucketCreator := storage.NewScratchBucketCreator(ctx, storageClient)
	computeClient, err := param.CreateComputeClient(&ctx, args.Oauth, args.ComputeEndpoint)
	if err != nil {
		return err
	}
	resourceLocationRetriever := storage.NewResourceLocationRetriever(metadataGCE, computeClient)

	region := new(string)
	paramPopulator := param.NewPopulator(
		param.NewNetworkResolver(computeClient),
		metadataGCE, storageClient,
		resourceLocationRetriever,
		scratchBucketCreator,
		param.NewMachineSeriesDetector(computeClient),
	)
	err = paramPopulator.PopulateMissingParameters(&args.Project, args.ClientID, &args.Zone, region, &args.ScratchBucketGcsPath,
		args.DestinationURI, nil, &args.Network, &args.Subnet, &args.WorkerMachineSeries)
	if err != nil {
		return err
	}

	var imageDiskSizeGb int64
	if args.SourceImage != "" {
		if imageDiskSizeGb, err = validateImageExists(computeClient, args.Project, args.SourceImage); err != nil {
			return err
		}
	} else {
		if imageDiskSizeGb, err = validateSnapshotExists(computeClient, args.Project, args.SourceDiskSnapshot); err != nil {
			return err
		}
	}

	varMap := buildDaisyVars(
		args.DestinationURI, args.SourceImage, args.SourceDiskSnapshot, imageDiskSizeGb,
		args.Format, args.Network, args.Subnet, *region, args.ComputeServiceAccount, args.QemuImgDockerImage)

	workflowProvider := func() (*daisy.Workflow, error) {
		wf, err := daisy.NewFromFile(getWorkflowPath(args.Format, args.CurrentExecutablePath))
		if err != nil {
			return nil, err
		}
		updateInstanceExternalIP(wf, args.ExternalIP)
		return wf, nil
	}

	env := daisyutils.EnvironmentSettings{
		Project:                     args.Project,
		Zone:                        args.Zone,
		GCSPath:                     args.ScratchBucketGcsPath,
		OAuth:                       args.Oauth,
		Timeout:                     args.Timeout,
		EndpointsOverride:           daisyutils.EndpointsOverride{Compute: args.ComputeEndpoint},
		DisableGCSLogs:              args.GcsLogsDisabled,
		DisableCloudLogs:            args.CloudLogsDisabled,
		DisableStdoutLogs:           args.StdoutLogsDisabled,
		Network:                     args.Network,
		Subnet:                      args.Subnet,
		ComputeServiceAccount:       args.ComputeServiceAccount,
		Labels:                      userLabels,
		ExecutionID:                 os.Getenv(os.Getenv(daisyutils.BuildIDOSEnvVarName)),
		WorkerMachineSeries:         args.WorkerMachineSeries,
		NestedVirtualizationEnabled: args.NestedVirtualizationEnabled,
		Tool: daisyutils.Tool{
			HumanReadableName: "gce image export",
			ResourceLabelName: "gce-image-export",
		},
	}

	if env.ExecutionID == "" {
		env.ExecutionID = path.RandString(5)
	}
	values, err := daisyutils.NewDaisyWorker(workflowProvider, env, logger).RunAndReadSerialValues(
		varMap, targetSizeGBKey, sourceSizeGBKey)
	logger.Metric(&pb.OutputInfo{
		SourcesSizeGb: []int64{stringutils.SafeStringToInt(values[sourceSizeGBKey])},
		TargetsSizeGb: []int64{stringutils.SafeStringToInt(values[targetSizeGBKey])},
	})
	return err
}

func updateInstanceExternalIP(wf *daisy.Workflow, externalIP string) {
	if externalIP == "" {
		return
	}
	for _, step := range wf.Steps {
		if step.IncludeWorkflow != nil && step.IncludeWorkflow.Workflow != nil {
			updateInstanceExternalIP(step.IncludeWorkflow.Workflow, externalIP)
			continue
		}
		if step.CreateInstances == nil {
			continue
		}

		if strings.EqualFold(externalIP, "none") {
			daisy.UpdateInstanceNoExternalIP(step)
			continue
		}

		targetIP := externalIP
		if strings.EqualFold(targetIP, "ephemeral") {
			targetIP = ""
		}

		for _, instance := range step.CreateInstances.Instances {
			for _, nic := range instance.Instance.NetworkInterfaces {
				if len(nic.AccessConfigs) == 0 {
					nic.AccessConfigs = []*computeV1.AccessConfig{{Type: "ONE_TO_ONE_NAT"}}
				}
				nic.AccessConfigs[0].NatIP = targetIP
			}
		}
	}
}

// validateImageExists checks whether imageName exists in the specified project.
//
// This validates when imageName is a valid image name, and skips validation if
// the imageName is a URI, or something that's not recognized as an image.
// The simplistic validation avoids false negatives; Daisy has robust logic for
// interpreting the various permutations of specifying an image and project,
// and we don't want to copy that here, since this is a convenience method to create
// user-friendly messages.
func validateImageExists(computeClient daisyCompute.Client, project string, imageURI string) (diskSizeGb int64, err error) {
	// try to get image even before validation in case it's a valid URL,
	// in order to obtain its size
	var imageName string
	if err := validation.ValidateImageName(imageURI); err != nil {
		if project, imageName, err = validation.ValidateImageURI(imageURI); err != nil {
			return diskSizeGb, nil
		}
	} else {
		imageName = imageURI
	}
	log.Printf("Fetching image %q from project %q.", imageName, project)
	image, err := computeClient.GetImage(project, imageName)
	if err != nil {
		log.Printf("Error when fetching image %q: %q.", imageURI, err)
		return diskSizeGb, daisy.Errf("Image %q not found", imageURI)
	}
	return image.DiskSizeGb, nil
}

// validateSnapshotExists checks whether snapshotName exists in the specified project.
//
// This validates when snapshotName is a valid snapshot name, and skips validation if
// the snapshotName is a URI, or something that's not recognized as a snapshot.
// The simplistic validation avoids false negatives; Daisy has robust logic for
// interpreting the various permutations of specifying a snapshot and project,
// and we don't want to copy that here, since this is a convenience method to create
// user-friendly messages.
func validateSnapshotExists(computeClient daisyCompute.Client, project string, snapshotName string) (diskSizeGb int64, err error) {
	// try to get snapshot even before validation in case it's a valid URL,
	// in order to obtain its size

	if err := validation.ValidateSnapshotName(snapshotName); err != nil {
		return diskSizeGb, nil
	}
	snapshot, err := computeClient.GetSnapshot(project, snapshotName)
	if err != nil {
		log.Printf("Error when fetching snapshot %q: %q.", snapshotName, err)
		return diskSizeGb, daisy.Errf("Snapshot %q not found", snapshotName)
	}

	return snapshot.DiskSizeGb, nil
}
