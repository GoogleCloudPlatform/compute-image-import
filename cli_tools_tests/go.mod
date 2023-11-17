module github.com/GoogleCloudPlatform/compute-image-import/cli_tools_tests

go 1.13

require (
	cloud.google.com/go/storage v1.31.0
	github.com/GoogleCloudPlatform/compute-daisy v0.0.0-20231114191308-36d2ee64eace
	github.com/GoogleCloudPlatform/compute-image-import/cli_tools v0.0.0
	github.com/GoogleCloudPlatform/compute-image-import/common v0.0.0
	github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils v0.0.0
	github.com/GoogleCloudPlatform/compute-image-import/proto/go v0.0.0
	github.com/aws/aws-sdk-go v1.37.5
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.8.3
	google.golang.org/api v0.129.0
	google.golang.org/protobuf v1.31.0
)

replace github.com/GoogleCloudPlatform/compute-image-import/common => ../common

replace github.com/GoogleCloudPlatform/compute-image-import/cli_tools => ../cli_tools

replace github.com/GoogleCloudPlatform/compute-image-import/go/e2e_test_utils => ../go/e2e_test_utils

replace github.com/GoogleCloudPlatform/compute-image-import/proto/go => ../proto/go
