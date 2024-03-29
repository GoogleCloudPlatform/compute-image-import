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

// Package exporttestsuites contains e2e tests for image export cli tools
package exporttestsuites

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"

	"github.com/GoogleCloudPlatform/compute-image-import/common/gcp"

	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/paramhelper"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/utils/path"
	"github.com/GoogleCloudPlatform/compute-image-import/cli_tools_tests/e2e"
	"github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils/junitxml"
	testconfig "github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils/test_config"
)

const (
	testSuiteName = "ImageExportTests"
)

var (
	argMap map[string]string
)

// TestSuite is image export test suite.
func TestSuite(ctx context.Context, tswg *sync.WaitGroup, testSuites chan *junitxml.TestSuite,
	logger *log.Logger, testSuiteRegex, testCaseRegex *regexp.Regexp, testProjectConfig *testconfig.Project, argMapInput map[string]string) {

	argMap = argMapInput

	testTypes := []e2e.CLITestType{
		e2e.Wrapper,
		e2e.GcloudBetaProdWrapperLatest,
		e2e.GcloudBetaLatestWrapperLatest,
		e2e.GcloudGaLatestWrapperRelease,
	}

	testsMap := map[e2e.CLITestType]map[*junitxml.TestCase]func(
		context.Context, *junitxml.TestCase, *log.Logger, *testconfig.Project, e2e.CLITestType){}

	for _, testType := range testTypes {
		imageExportRawTestCase := junitxml.NewTestCase(
			testSuiteName, fmt.Sprintf("[%v] %v", testType, "Raw"))
		imageExportVMDKTestCase := junitxml.NewTestCase(
			testSuiteName, fmt.Sprintf("[%v] %v", testType, "VMDK"))
		imageExportWithRichParamsTestCase := junitxml.NewTestCase(
			testSuiteName, fmt.Sprintf("[%v] %v", testType, "With rich params"))
		imageExportWithDifferentNetworkParamStyles := junitxml.NewTestCase(
			testSuiteName, fmt.Sprintf("[%v] %v", testType, "With different network param styles"))
		imageExportWithSubnetWithoutNetworkTestCase := junitxml.NewTestCase(
			testSuiteName, fmt.Sprintf("[%v] %v", testType, "With subnet but without network"))

		testsMap[testType] = map[*junitxml.TestCase]func(
			context.Context, *junitxml.TestCase, *log.Logger, *testconfig.Project, e2e.CLITestType){}
		testsMap[testType][imageExportRawTestCase] = runImageExportRawTest
		testsMap[testType][imageExportVMDKTestCase] = runImageExportVMDKTest
		testsMap[testType][imageExportWithRichParamsTestCase] = runImageExportWithRichParamsTest
		testsMap[testType][imageExportWithDifferentNetworkParamStyles] = runImageExportWithDifferentNetworkParamStyles
		testsMap[testType][imageExportWithSubnetWithoutNetworkTestCase] = runImageExportWithSubnetWithoutNetworkParamsTest
	}

	// Only test service account scenario for wrapper, until gcloud support it.
	imageExportRawWithoutDefaultServiceAccountTestCase := junitxml.NewTestCase(
		testSuiteName, fmt.Sprintf("[%v] %v", e2e.Wrapper, "Raw without default service account"))
	imageExportVMDKDefaultServiceAccountWithMissingPermissionsTestCase := junitxml.NewTestCase(
		testSuiteName, fmt.Sprintf("[%v] %v", e2e.Wrapper, "VMDK without default service account permission"))
	testsMap[e2e.Wrapper][imageExportRawWithoutDefaultServiceAccountTestCase] = runImageExportRawWithoutDefaultServiceAccountTest
	testsMap[e2e.Wrapper][imageExportVMDKDefaultServiceAccountWithMissingPermissionsTestCase] = runImageExportVMDKDefaultServiceAccountWithMissingPermissionsTest

	e2e.CLITestSuite(ctx, tswg, testSuites, logger, testSuiteRegex, testCaseRegex, testProjectConfig, testSuiteName, testsMap)
}

func runImageExportRawTest(ctx context.Context, testCase *junitxml.TestCase, logger *log.Logger,
	testProjectConfig *testconfig.Project, testType e2e.CLITestType) {

	suffix := path.RandString(5)
	bucketName := fmt.Sprintf("%v-test-image-eu", testProjectConfig.TestProjectID)
	objectName := fmt.Sprintf("e2e-export-raw-test-%v", suffix)
	fileURI := fmt.Sprintf("gs://%v/%v", bucketName, objectName)
	zone := "europe-west1-c"

	argsMap := map[e2e.CLITestType][]string{
		e2e.Wrapper: {"-client_id=e2e", fmt.Sprintf("-project=%v", testProjectConfig.TestProjectID),
			"-source_image=global/images/e2e-test-image-10g-eu", fmt.Sprintf("-destination_uri=%v", fileURI),
			fmt.Sprintf("-zone=%v", zone),
			"-worker_machine_series=n1",
		},
		e2e.GcloudBetaProdWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID), "--image=e2e-test-image-10g-eu",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudBetaLatestWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID), "--image=e2e-test-image-10g-eu",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudGaLatestWrapperRelease: {"compute", "images", "export", "--quiet",
			fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID), "--image=e2e-test-image-10g-eu",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
	}

	runExportTest(ctx, argsMap[testType], testType, logger, testCase, bucketName, objectName)
}

func runImageExportVMDKTest(ctx context.Context, testCase *junitxml.TestCase, logger *log.Logger,
	testProjectConfig *testconfig.Project, testType e2e.CLITestType) {

	suffix := path.RandString(5)
	bucketName := fmt.Sprintf("%v-test-image-asia", testProjectConfig.TestProjectID)
	objectName := fmt.Sprintf("e2e-export-vmdk-test-%v", suffix)
	fileURI := fmt.Sprintf("gs://%v/%v", bucketName, objectName)
	zone := "asia-northeast1-a"

	argsMap := map[e2e.CLITestType][]string{
		e2e.Wrapper: {"-client_id=e2e", fmt.Sprintf("-project=%v", testProjectConfig.TestProjectID),
			"-source_image=global/images/e2e-test-image-10g-asia", fmt.Sprintf("-destination_uri=%v", fileURI), "-format=vmdk",
			fmt.Sprintf("-zone=%v", zone),
			"-worker_machine_series=n1",
		},
		e2e.GcloudBetaProdWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID), "--image=e2e-test-image-10g-asia",
			fmt.Sprintf("--destination-uri=%v", fileURI), "--export-format=vmdk", fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudBetaLatestWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID), "--image=e2e-test-image-10g-asia",
			fmt.Sprintf("--destination-uri=%v", fileURI), "--export-format=vmdk", fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudGaLatestWrapperRelease: {"compute", "images", "export", "--quiet",
			fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID), "--image=e2e-test-image-10g-asia",
			fmt.Sprintf("--destination-uri=%v", fileURI), "--export-format=vmdk", fmt.Sprintf("--zone=%v", zone),
		},
	}

	runExportTest(ctx, argsMap[testType], testType, logger, testCase, bucketName, objectName)
}

// Test most of params except -oauth, -compute_endpoint_override, and -scratch_bucket_gcs_path
func runImageExportWithRichParamsTest(ctx context.Context, testCase *junitxml.TestCase,
	logger *log.Logger, testProjectConfig *testconfig.Project, testType e2e.CLITestType) {

	suffix := path.RandString(5)
	bucketName := fmt.Sprintf("%v-test-image", testProjectConfig.TestProjectID)
	objectName := fmt.Sprintf("e2e-export-rich-param-test-%v", suffix)
	fileURI := fmt.Sprintf("gs://%v/%v", bucketName, objectName)
	zone := "us-central1-c" //has to be in us-central1 as subnets are configured there

	argsMap := map[e2e.CLITestType][]string{
		e2e.Wrapper: {"-client_id=e2e", fmt.Sprintf("-project=%v", testProjectConfig.TestProjectID),
			"-source_image=global/images/e2e-test-image-10g", fmt.Sprintf("-destination_uri=%v", fileURI),
			fmt.Sprintf("-network=%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("-subnet=%v-subnet-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("-zone=%v", zone),
			"-timeout=2h", "-disable_gcs_logging", "-disable_cloud_logging", "-disable_stdout_logging",
			"-labels=key1=value1,key2=value",
			"-worker_machine_series=n1",
		},
		e2e.GcloudBetaProdWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--network=%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=%v-subnet-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--zone=%v", zone),
			"--timeout=2h", "--image=e2e-test-image-10g", fmt.Sprintf("--destination-uri=%v", fileURI),
		},
		e2e.GcloudBetaLatestWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--network=%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=%v-subnet-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--zone=%v", zone),
			"--timeout=2h", "--image=e2e-test-image-10g", fmt.Sprintf("--destination-uri=%v", fileURI),
		},
		e2e.GcloudGaLatestWrapperRelease: {"compute", "images", "export", "--quiet",
			fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--network=%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=%v-subnet-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--zone=%v", zone),
			"--timeout=2h", "--image=e2e-test-image-10g", fmt.Sprintf("--destination-uri=%v", fileURI),
		},
	}

	runExportTest(ctx, argsMap[testType], testType, logger, testCase, bucketName, objectName)
}

func runImageExportWithDifferentNetworkParamStyles(ctx context.Context, testCase *junitxml.TestCase,
	logger *log.Logger, testProjectConfig *testconfig.Project, testType e2e.CLITestType) {

	suffix := path.RandString(5)
	bucketName := fmt.Sprintf("%v-test-image", testProjectConfig.TestProjectID)
	objectName := fmt.Sprintf("e2e-export-subnet-test-%v", suffix)
	fileURI := fmt.Sprintf("gs://%v/%v", bucketName, objectName)
	zone := "us-central1-c" //has to be in us-central1 as subnets are configured there
	region, _ := paramhelper.GetRegion(testProjectConfig.TestZone)

	argsMap := map[e2e.CLITestType][]string{
		e2e.Wrapper: {"-client_id=e2e", fmt.Sprintf("-project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("-network=global/networks/%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("-subnet=projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"-source_image=global/images/e2e-test-image-10g", fmt.Sprintf("-destination_uri=%v", fileURI),
			fmt.Sprintf("-zone=%v", zone),
			"-worker_machine_series=n1",
		},
		e2e.GcloudBetaProdWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--network=global/networks/%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"--image=e2e-test-image-10g",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudBetaLatestWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--network=global/networks/%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"--image=e2e-test-image-10g",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudGaLatestWrapperRelease: {"compute", "images", "export", "--quiet",
			fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--network=global/networks/%v-vpc-1", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"--image=e2e-test-image-10g",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
	}

	runExportTest(ctx, argsMap[testType], testType, logger, testCase, bucketName, objectName)
}

func runImageExportWithSubnetWithoutNetworkParamsTest(ctx context.Context, testCase *junitxml.TestCase,
	logger *log.Logger, testProjectConfig *testconfig.Project, testType e2e.CLITestType) {

	suffix := path.RandString(5)
	bucketName := fmt.Sprintf("%v-test-image", testProjectConfig.TestProjectID)
	objectName := fmt.Sprintf("e2e-export-subnet-test-%v", suffix)
	fileURI := fmt.Sprintf("gs://%v/%v", bucketName, objectName)
	zone := "us-central1-c" //has to be us-central-1 as subnets are configured in that region
	region, _ := paramhelper.GetRegion(zone)

	argsMap := map[e2e.CLITestType][]string{
		e2e.Wrapper: {"-client_id=e2e", fmt.Sprintf("-project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("-subnet=https://www.googleapis.com/compute/v1/projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"-source_image=global/images/e2e-test-image-10g", fmt.Sprintf("-destination_uri=%v", fileURI),
			fmt.Sprintf("-zone=%v", zone),
			"-worker_machine_series=n1",
		},
		e2e.GcloudBetaProdWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=https://www.googleapis.com/compute/v1/projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"--image=e2e-test-image-10g",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudBetaLatestWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=https://www.googleapis.com/compute/v1/projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"--image=e2e-test-image-10g",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
		e2e.GcloudGaLatestWrapperRelease: {"compute", "images", "export", "--quiet",
			fmt.Sprintf("--project=%v", testProjectConfig.TestProjectID),
			fmt.Sprintf("--subnet=https://www.googleapis.com/compute/v1/projects/%v/regions/%v/subnetworks/%v-subnet-1",
				testProjectConfig.TestProjectID, region, testProjectConfig.TestProjectID),
			"--image=e2e-test-image-10g",
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", zone),
		},
	}

	runExportTest(ctx, argsMap[testType], testType, logger, testCase, bucketName, objectName)
}

func runImageExportRawWithoutDefaultServiceAccountTest(ctx context.Context, testCase *junitxml.TestCase, logger *log.Logger,
	testProjectConfig *testconfig.Project, testType e2e.CLITestType) {

	testVariables, ok := e2e.GetServiceAccountTestVariables(argMap, true)
	if !ok {
		e2e.Failure(testCase, logger, fmt.Sprintln("Failed to get service account test args"))
		return
	}

	suffix := path.RandString(5)
	bucketName := fmt.Sprintf("%v-test-image", testProjectConfig.TestProjectID)
	objectName := fmt.Sprintf("e2e-export-raw-test-%v", suffix)
	fileURI := fmt.Sprintf("gs://%v/%v", bucketName, objectName)

	argsMap := map[e2e.CLITestType][]string{
		e2e.Wrapper: {"-client_id=e2e", fmt.Sprintf("-project=%v", testVariables.ProjectID),
			fmt.Sprintf("-source_image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("-destination_uri=%v", fileURI),
			fmt.Sprintf("-zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("-compute_service_account=%v", testVariables.ComputeServiceAccount),
			"-worker_machine_series=n1",
		},
		e2e.GcloudBetaProdWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testVariables.ProjectID),
			fmt.Sprintf("--image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("--compute-service-account=%v", testVariables.ComputeServiceAccount),
		},
		e2e.GcloudBetaLatestWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testVariables.ProjectID),
			fmt.Sprintf("--image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("--compute-service-account=%v", testVariables.ComputeServiceAccount),
		},
		e2e.GcloudGaLatestWrapperRelease: {"compute", "images", "export", "--quiet",
			fmt.Sprintf("--project=%v", testVariables.ProjectID),
			fmt.Sprintf("--image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("--destination-uri=%v", fileURI), fmt.Sprintf("--zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("--compute-service-account=%v", testVariables.ComputeServiceAccount),
		},
	}

	runExportTest(ctx, argsMap[testType], testType, logger, testCase, bucketName, objectName)
}

func runImageExportVMDKDefaultServiceAccountWithMissingPermissionsTest(ctx context.Context, testCase *junitxml.TestCase, logger *log.Logger,
	testProjectConfig *testconfig.Project, testType e2e.CLITestType) {

	testVariables, ok := e2e.GetServiceAccountTestVariables(argMap, false)
	if !ok {
		e2e.Failure(testCase, logger, fmt.Sprintln("Failed to get service account test args"))
		return
	}

	suffix := path.RandString(5)
	bucketName := fmt.Sprintf("%v-test-image", testProjectConfig.TestProjectID)
	objectName := fmt.Sprintf("e2e-export-vmdk-test-%v", suffix)
	fileURI := fmt.Sprintf("gs://%v/%v", bucketName, objectName)

	argsMap := map[e2e.CLITestType][]string{
		e2e.Wrapper: {"-client_id=e2e", fmt.Sprintf("-project=%v", testVariables.ProjectID),
			fmt.Sprintf("-source_image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("-destination_uri=%v", fileURI), "-format=vmdk",
			fmt.Sprintf("-zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("-compute_service_account=%v", testVariables.ComputeServiceAccount),
			"-worker_machine_series=n1",
		},
		e2e.GcloudBetaProdWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testVariables.ProjectID),
			fmt.Sprintf("--image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("--destination-uri=%v", fileURI), "--export-format=vmdk", fmt.Sprintf("--zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("--compute-service-account=%v", testVariables.ComputeServiceAccount),
		},
		e2e.GcloudBetaLatestWrapperLatest: {"beta", "compute", "images", "export", "--quiet",
			"--docker-image-tag=latest", fmt.Sprintf("--project=%v", testVariables.ProjectID),
			fmt.Sprintf("--image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("--destination-uri=%v", fileURI), "--export-format=vmdk", fmt.Sprintf("--zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("--compute-service-account=%v", testVariables.ComputeServiceAccount),
		},
		e2e.GcloudGaLatestWrapperRelease: {"compute", "images", "export", "--quiet",
			fmt.Sprintf("--project=%v", testVariables.ProjectID),
			fmt.Sprintf("--image=projects/%v/global/images/e2e-test-image-10g", testProjectConfig.TestProjectID),
			fmt.Sprintf("--destination-uri=%v", fileURI), "--export-format=vmdk", fmt.Sprintf("--zone=%v", testProjectConfig.TestZone),
			fmt.Sprintf("--compute-service-account=%v", testVariables.ComputeServiceAccount),
		},
	}

	runExportTest(ctx, argsMap[testType], testType, logger, testCase, bucketName, objectName)
}

func runExportTest(ctx context.Context, args []string, testType e2e.CLITestType,
	logger *log.Logger, testCase *junitxml.TestCase, bucketName string, objectName string) {

	cmds := map[e2e.CLITestType]string{
		e2e.Wrapper:                       "./gce_vm_image_export",
		e2e.GcloudBetaProdWrapperLatest:   "gcloud",
		e2e.GcloudBetaLatestWrapperLatest: "gcloud",
		e2e.GcloudGaLatestWrapperRelease:  "gcloud",
	}

	if e2e.RunTestForTestType(cmds[testType], args, testType, logger, testCase) {
		verifyExportedImageFile(ctx, testCase, bucketName, objectName, logger)
	}
}

func verifyExportedImageFile(ctx context.Context, testCase *junitxml.TestCase, bucketName string,
	objectName string, logger *log.Logger) {

	logger.Printf("Verifying exported file...")
	file, err := gcp.CreateFileObject(ctx, bucketName, objectName)
	if err != nil {
		testCase.WriteFailure("File '%v' doesn't exist after export: %v", objectName, err)
		logger.Printf("File '%v' doesn't exist after export: %v", objectName, err)
		return
	}
	logger.Printf("File '%v' exists! Export success.", objectName)

	if err := file.Cleanup(); err != nil {
		logger.Printf("File '%v' failed to clean up.", objectName)
	} else {
		logger.Printf("File '%v' cleaned up.", objectName)
	}
}
