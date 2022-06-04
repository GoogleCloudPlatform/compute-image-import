# Debian 11 Worker Build workflow

### debian_11_worker.wf.json

Imports a virtual disk file and converts it into a GCE image resource.

Variables:
* `build_date`: Build datestamp used to version the image. Default: Current date in format YYYYMMDD.
* `family-tag`: Image family name used as a base image. Default: debian-11
* `image_prefix`: Prefix for the created image. Default: debian-11-worker

Example Daisy invocation:
```shell
# Example importing a VMDK (using a credentials file)
daisy -project my-project \
      -gcs_path gs://bucket/daisyscratch \
      debian_11_worker.wf.json
```
