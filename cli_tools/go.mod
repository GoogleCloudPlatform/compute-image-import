module github.com/GoogleCloudPlatform/compute-image-import/cli_tools

go 1.13

require (
	cloud.google.com/go/compute/metadata v0.2.3
	cloud.google.com/go/storage v1.30.1
	cos.googlesource.com/cos/tools.git v0.0.0-20210104210903-4b3bc7d49b79 // indirect
	github.com/GoogleCloudPlatform/compute-daisy v0.0.0-20220223233810-60345cd7065c
	github.com/GoogleCloudPlatform/compute-image-import/common v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/compute-image-import/proto/go v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/osconfig v0.0.0-20210202205636-8f5a30e8969f
	github.com/aws/aws-sdk-go v1.37.5
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/go-playground/validator/v10 v10.4.1
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.9
	github.com/google/logger v1.1.0
	github.com/google/uuid v1.3.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/minio/highwayhash v1.0.1
	github.com/stretchr/testify v1.8.1
	github.com/vmware/govmomi v0.24.0
	go.chromium.org/luci v0.0.0-20210204234011-34a994fe5aec // indirect
	golang.org/x/sync v0.1.0
	golang.org/x/sys v0.6.0
	google.golang.org/api v0.114.0
	google.golang.org/protobuf v1.31.0
)

replace github.com/GoogleCloudPlatform/compute-image-import/proto/go => ../proto/go

replace github.com/GoogleCloudPlatform/compute-image-import/common => ../common
