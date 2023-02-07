## Image Import and Export Tools

This folder contains various tools for managing disk images on Google
Compute Engine.

### VM Image Import

The `gce_vm_image_import` tool imports a VM image to Google Compute Engine.
It uses Daisy to perform imports while adding additional logic to perform
import setup and clean-up, such as creating a temporary bucket, validating
flags etc.

### OVF Image Import

The `gce_ovf_import` tool imports a virtual appliance in OVF format 
to a Google Compute Engine VM or a Google Compute Engine
machine image.

### VM Image Export

The `gce_vm_image_export` tool exports a VM image, or a disk snapshot to Google Cloud Storage.
It uses Daisy to perform exports while adding additional logic to perform
export setup and clean-up, such as validating flags.

### OVF Image Export

The `gce_ovf_export` tool exports a Google Compute Engine VM or a Google Compute
Engine machine image to a virtual appliance in OVF format.

### Image Import Precheck Tool

The `import_precheck` tool runs on your VM image before attempting to import it into
Google Compute Engine and identifies compatibility issues that
will either cause import to fail or will cause potentially unexpected behavior
after import.

### One-step Image Import

The `gce_onestep_image_import` tool imports a VM image from other cloud providers to Google Compute Engine.
