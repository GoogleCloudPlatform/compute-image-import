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

package main

import (
	"context"
	"log"
	"os"
	"regexp"
	"sync"

	instanceovfexporttestsuite "github.com/GoogleCloudPlatform/compute-image-import/cli_tools_tests/e2e/gce_ovf_export/test_suites/instance_ovf_export"
	e2etestutils "github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils"
	"github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils/junitxml"
	testconfig "github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils/test_config"
)

func main() {
	instanceOvfExportTestSuccess := e2etestutils.RunTestsAndOutput([]func(context.Context, *sync.WaitGroup, chan *junitxml.TestSuite, *log.Logger,
		*regexp.Regexp, *regexp.Regexp, *testconfig.Project){instanceovfexporttestsuite.TestSuite},
		"[OVFInstanceExportTests]")
	if !instanceOvfExportTestSuccess {
		os.Exit(1)
	}
}
