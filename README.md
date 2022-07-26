# Compute Engine Image Import

This repository contains various tools for managing disk images on Google
Compute Engine.

## Docs

The main documentation for the tools in this repository can be found on our
[GitHub.io page](https://googlecloudplatform.github.io/compute-image-import/image-import.html).

## Daisy

Daisy is a solution for running multi-step workflows on GCE.

### Daisy Workflows

Full featured Daisy workflow examples for image import
workflows. A [user guide](https://googlecloudplatform.github.io/compute-image-import/image-import.html) for VM imports is
also provided here.

## GCE Export tool

The gce_ovf_export and gce_vm_image_export tools stream a local disk to a Google Compute Engine image file in a Google Cloud Storage bucket.

### Prebuilt binaries
Prebuilt binaries are available for Windows, and Linux.

Built from the latest GitHub release (all 64bit):

+ [Windows](https://storage.googleapis.com/compute-image-tools/release/windows/gce_export.exe)
+ [Linux](https://storage.googleapis.com/compute-image-tools/release/linux/gce_export)

Built from the latest commit to the master branch (all 64bit):

+ [Windows](https://storage.googleapis.com/compute-image-tools/latest/windows/gce_export.exe)
+ [Linux](https://storage.googleapis.com/compute-image-tools/latest/linux/gce_export)

## Contributing

Have a patch that will benefit this project? Awesome! Follow these steps to have
it accepted.

1.  Please sign our [Contributor License Agreement](CONTRIBUTING.md).
1.  Fork this Git repository and make your changes.
1.  Create a Pull Request.
1.  Incorporate review feedback to your changes.
1.  Accepted!

## License

All files in this repository are under the
[Apache License, Version 2.0](LICENSE) unless noted otherwise.
