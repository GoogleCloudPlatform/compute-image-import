module github.com/GoogleCloudPlatform/compute-image-import/cli_tools

go 1.21.5

require (
	cloud.google.com/go/compute/metadata v0.2.3
	cloud.google.com/go/logging v1.7.0
	cloud.google.com/go/storage v1.31.0
	github.com/GoogleCloudPlatform/compute-daisy v0.0.0-20231114191308-36d2ee64eace
	github.com/GoogleCloudPlatform/compute-image-import/common v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/compute-image-import/proto/go v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/osconfig v0.0.0-20210202205636-8f5a30e8969f
	github.com/aws/aws-sdk-go v1.37.5
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.9
	github.com/google/logger v1.1.0
	github.com/google/uuid v1.3.0
	github.com/minio/highwayhash v1.0.1
	github.com/stretchr/testify v1.8.3
	github.com/vmware/govmomi v0.24.0
	golang.org/x/sync v0.3.0
	golang.org/x/sys v0.18.0
	google.golang.org/api v0.129.0
	google.golang.org/protobuf v1.31.0
)

require (
	cloud.google.com/go v0.110.3 // indirect
	cloud.google.com/go/compute v1.20.1 // indirect
	cloud.google.com/go/iam v1.1.1 // indirect
	cloud.google.com/go/longrunning v0.5.1 // indirect
	cos.googlesource.com/cos/tools.git v0.0.0-20210104210903-4b3bc7d49b79 // indirect
	github.com/GoogleCloudPlatform/guest-logging-go v0.0.0-20200113214433-6cbb518174d4 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/glog v1.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.5 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	go.chromium.org/luci v0.0.0-20210204234011-34a994fe5aec // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/oauth2 v0.9.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230629202037-9506855d4529 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230629202037-9506855d4529 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230629202037-9506855d4529 // indirect
	google.golang.org/grpc v1.56.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/GoogleCloudPlatform/compute-image-import/proto/go => ../proto/go

replace github.com/GoogleCloudPlatform/compute-image-import/common => ../common
