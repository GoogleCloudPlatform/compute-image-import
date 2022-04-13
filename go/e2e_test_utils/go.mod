module github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils

go 1.13

require (
	github.com/GoogleCloudPlatform/compute-image-import/cli_tools v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.2.0
)

replace github.com/GoogleCloudPlatform/compute-image-import/cli_tools => ../../cli_tools

replace github.com/GoogleCloudPlatform/compute-image-import/proto/go => ../../proto/go

replace github.com/GoogleCloudPlatform/compute-image-import/common => ../../common
