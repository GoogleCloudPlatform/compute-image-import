# Debian Worker Build workflow

### debian_worker.wf.json

Build a new Debian worker image for GCE import/export operations.
This workflow has been tested with Debian 10 and 11.

Variables:
* `build_tag`: Build tag used to version the image. Default: Current date in format YYYYMMDD.
* `commit_sha`: Git commit hash ($COMMIT_SHA in Cloud Build) to link the worker build to its commit. Default: ""
* `family_tag`: Image family name used as a base image. Default: debian-11
* `image_prefix`: Prefix for the created image. Default: debian-11-worker
* `source_image`: Source image for Debian worker. Default: projects/debian-cloud/global/images/family/debian-11

Example Daisy invocation:
```shell
# Example of building Debian 11 worker
daisy -project my-project \
      -gcs_path gs://bucket/daisyscratch \
      debian_worker.wf.json
```
