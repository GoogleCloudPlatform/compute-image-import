# Importing Virtual Disks into Google Compute Engine (GCE)

Note: Google Compute Engine supports importing virtual disks and virtual appliances
by using the **image import** tool. For more information, see the
[import tool documentation](https://cloud.google.com/compute/docs/import/importing-virtual-disks)

Some basic concepts to start with:

*   **Virtual Disk**: Virtual disk is a file that encapsulates the content of a virtualized disk in a virtualization environment. Virtual Disks are critical components of virtual machines for holding boot media and data. Virtualization platforms (e.g. VMWare, Hyper-v, KVM, etc.) each have their own format for virtual disks.
*   **Persistent Disk**: Compute Engine Persistent Disk is a Compute Engine resource that is equivalent to disk drives in physical computers and virtualdisks in a virtualization environment.
*   **GCE Image**: Image is an immutable representation of a Persistent Disk and is used for creating multiple disks from one single templatized version.

**NOTE:** Before attempting a virtual disk import, take a look at the [known compatibility issues](#compatibility-and-known-limitations) and our [compatibility precheck tool](#compatibility-precheck-tool) below.

## Compatibility and Known Limitations

*   Networking: Import workflow sets the interface to DHCP. If that fails, or if there are other interfaces set with firewalls, special routing, VPN's, or other non-standard configurations, networking may fail and while the resulting instance may boot, you may not be able to access it.

Not every VM image will be importable to GCE. Some VMs will have issues after
import. Below is a list of known compatibility requirements and issues:

## Compatibility Precheck Tool
Image import has a long runtime, can fail due to incompatibilities, and can
cause unexpected behavior post-import. As such, you may find it useful to run
our [precheck tool](https://github.com/GoogleCloudPlatform/compute-image-import/tree/master/cli_tools/import_precheck/)
to check for the known issues listed above.