# Compute Engine Image Import

This repository contains various tools for managing disk images on Google
Compute Engine using Daisy.

**Note:** Google no longer provides support for the import of virtual disks and appliances using this tool,
as it reached EOL in July 2025. 
Please use [M2VM Image Import](https://cloud.google.com/migrate/virtual-machines/docs/5.0/migrate/image_import) for virtual disks and [M2VM Machine Image Import](https://cloud.google.com/migrate/virtual-machines/docs/5.0/migrate/machine-image-import) for virtual appliances.

The documentation for the tools in this repository can be found on our
[GitHub.io page](https://googlecloudplatform.github.io/compute-image-import/image-import.html).

## Daisy

Daisy is a solution for running multi-step workflows on Google Compute Engine.

### Daisy Workflows

This repository contains full featured Daisy workflow examples for image import.
A user guide for importing virtual disk using the Daisy workflow is
[here](https://googlecloudplatform.github.io/compute-image-import/image-import.html).

## Image Import and Export Tools

The `cli_tools` folder in this repository contains the tools for importing and exporting images using Daisy.

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
