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

package validation

import (
	"fmt"
	"regexp"
	"strings"

	daisy "github.com/GoogleCloudPlatform/compute-daisy"
)

const (
	rfc1035LabelRegexpStr = "[A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9]"
	imageNameStr          = "[a-z]([-a-z0-9]{0,61}[a-z0-9])?"
	diskSnapshotNameStr   = "^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$"

	// projectID: "The unique, user-assigned ID of the Project. It must be 6 to 30
	// lowercase letters,  digits, or hyphens. It must start with a letter.
	// Trailing hyphens are prohibited."
	// -- https://cloud.google.com/resource-manager/reference/rest/v1/projects
	projectIDStr = "(google\\.com\\:)?[a-z][-a-z0-9]{4,28}[a-z0-9]" //
)

var (
	rfc1035LabelRegexp     = regexp.MustCompile(rfc1035LabelRegexpStr)
	fqdnRegexp             = regexp.MustCompile(fmt.Sprintf("^((%v)\\.)+(%v)$", rfc1035LabelRegexpStr, rfc1035LabelRegexpStr))
	imageNameRegexp        = regexp.MustCompile(fmt.Sprintf("^%v$", imageNameStr))
	imageURIRegexp         = regexp.MustCompile(fmt.Sprintf("^projects/(%v)/global/images/(%v)$", projectIDStr, imageNameStr))
	diskSnapshotNameRegexp = regexp.MustCompile(diskSnapshotNameStr)
	projectIDRegexp        = regexp.MustCompile(fmt.Sprintf("^%v$", projectIDStr))
)

// ValidateStringFlagNotEmpty returns error with error message stating field must be provided if
// value is empty string. Returns nil otherwise.
func ValidateStringFlagNotEmpty(flagValue string, flagKey string) error {
	if flagValue == "" {
		return daisy.Errf(fmt.Sprintf("The flag -%v must be provided", flagKey))
	}
	return nil
}

// ValidateExactlyOneOfStringFlagNotEmpty returns error with error message stating one of fields must be provided if
// value is empty string. Returns nil otherwise.
func ValidateExactlyOneOfStringFlagNotEmpty(flagKeyValues map[string]string) error {
	var notEmpty []string
	for k, v := range flagKeyValues {
		if v != "" {
			notEmpty = append(notEmpty, k)
		}
	}
	if len(notEmpty) != 1 {
		return daisy.Errf(fmt.Sprintf("Exactly one of -%v flags should be provided", strings.Join(notEmpty, ",-")))
	}
	return nil
}

// ValidateFqdn validates fully qualified domain name
func ValidateFqdn(flagValue string, flagKey string) error {
	flagValueLen := len(flagValue)
	if flagValueLen < 1 || flagValueLen > 253 || !fqdnRegexp.MatchString(flagValue) {
		return daisy.Errf(fmt.Sprintf("The flag `%v` must conform to RFC 1035 requirements for valid hostnames. "+
			"To meet this requirement, the value must contain a series of labels and each label is concatenated with a dot. "+
			"Each label can be 1-63 characters long where each character can be a letter, a digit or a dash (`-`). The "+
			"entire sequence must not exceed 253 characters.", flagKey))
	}
	return nil
}

// ValidateRfc1035Label validates a single label per RFC 1035
func ValidateRfc1035Label(value string) error {
	if len(value) > 63 || !rfc1035LabelRegexp.MatchString(value) {
		return daisy.Errf(fmt.Sprintf("Value `%v` must conform to RFC 1035 requirements for valid labels.", value))
	}
	return nil
}

// ValidateImageName validates whether a string is a valid image name, as defined by
// <https://cloud.google.com/compute/docs/reference/rest/v1/images>.
func ValidateImageName(value string) error {
	if !imageNameRegexp.MatchString(value) {
		return daisy.Errf("Image name `%v` must conform to https://cloud.google.com/compute/docs/reference/rest/v1/images", value)
	}
	return nil
}

// ValidateImageURI validates whether a string is a valid image URI, as defined by
// <https://cloud.google.com/compute/docs/reference/rest/v1/images> and returns
// image name and project ID if valid.
func ValidateImageURI(value string) (project string, imageName string, err error) {
	if !imageURIRegexp.MatchString(value) {
		return imageName, project, daisy.Errf("Image URI `%v` must conform to https://cloud.google.com/compute/docs/reference/rest/v1/images", value)
	}
	match := imageURIRegexp.FindStringSubmatch(value)
	return match[1], match[3], nil
}

// ValidateSnapshotName validates whether a string is a valid disk snapshot name, as defined by
// <https://cloud.google.com/compute/docs/reference/rest/v1/snapshots>.
func ValidateSnapshotName(value string) error {
	if !diskSnapshotNameRegexp.MatchString(value) {
		return daisy.Errf("Snapshot name `%v` must conform to https://cloud.google.com/compute/docs/reference/rest/v1/snapshots", value)
	}
	return nil
}

// ValidateProjectID validates whether a string is a valid projectID, as defined by
// <https://cloud.google.com/resource-manager/reference/rest/v1/projects>.
func ValidateProjectID(value string) error {
	if !projectIDRegexp.MatchString(value) {
		return daisy.Errf("projectID `%v` must conform to https://cloud.google.com/resource-manager/reference/rest/v1/projects", value)
	}
	return nil
}
