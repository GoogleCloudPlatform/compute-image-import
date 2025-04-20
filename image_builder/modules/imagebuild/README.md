## Overview


## Usage

Basic usage of this module is as follows:

```hcl
module "image-pipeline" {
  source = "./modules/imagebuild"

  project_id           = <PROJECT_ID>
  pipeline_name        = "imagebuild-pipeline"
  region               = "us-central1"
  zone                 = "us-central1-c"
  gcs_folder           = <GCS_PATH_FOR_SCRIPT_STORAGE>
  source_image = ({
    image = "projects/debian-cloud/global/images/debian-12-bookworm-v20250311"
  })
  target_image_name           = "sample1"
  target_image_region         = "us-east1"
  service_account_id          = "<SERVICE_ACCOUNT_FULL_ID>"
  customization_script_source = "scripts/customize.pkr.hcl"
  testing_script_source       = "scripts/test.pkr.hcl"
}
```

## Features

1. Create a new image building pipeline
1. Update customization and test scripts
1. Set a pipeline schedule



## Resources created

- Cloudbuild trigger
- (optional) GCS objects for customization & test scripts
- (optional) Scheduler job for triggering on a schedule

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| customization\_script\_source | Path to local customize script. | `string` | `""` | no |
| gcs\_folder | Folder in which customization scripts are located. | `string` | n/a | yes |
| pipeline\_description | Description attached to the cloudbuild trigger. | `string` | `""` | no |
| pipeline\_name | Name for of the Image build. Cloudbuild trigger will have this name. | `string` | n/a | yes |
| project\_id | Project ID for Image build Pipeline Project. | `string` | n/a | yes |
| region | Region used for Cloudbuild trigger. | `string` | n/a | yes |
| schedule\_cron\_pattern | Optional cron pattern for setting up a schedule. | `string` | `""` | no |
| schedule\_timezone | Optional timezone for the schedule. Must be a time zone name from the tz database. | `string` | `"UTC"` | no |
| service\_account\_id | ID of the service account to use. | `string` | n/a | yes |
| source\_image | Initial image to customize, can be one of:<br/>  - image: Full path of image to use.<br/>  - image\_family: Full path of image family to use.<br/>  - marketplace\_image: Full path of market place image.<br/>  - import: Object describing an image import from a GCS file. | <pre>object({<br/>    image             = optional(string)<br/>    image_family      = optional(string)<br/>    marketplace_image = optional(string)<br/>    import = optional(object({<br/>      gcs_path           = string # imagebuilder-import-source/vmware_ubuntu22.04_ubuntu22.04-server.vmdk<br/>      license_type       = optional(string)<br/>      generalize         = bool<br/>      skip_os_adaptation = bool<br/>    }))<br/>  })</pre> | n/a | yes |
| target\_image\_description | Description attached to the target image created. | `string` | `""` | no |
| target\_image\_encryption\_key | CMEK used for target image (and processing). | `string` | `""` | no |
| target\_image\_family | Family for target image (optional). | `string` | `""` | no |
| target\_image\_labels | Key-value pairs of additional labels to attach to the target image. | `map(string)` | `{}` | no |
| target\_image\_name | Name for target image. | `string` | n/a | yes |
| target\_image\_region | Region used for target image. | `string` | `""` | no |
| testing\_script\_source | Path to local test script. | `string` | `""` | no |
| zone | Zone used for cloud build instances. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| trigger\_id | CloudBuild trigger ID. |
| trigger\_name | CloudBuild trigger Name. |
| trigger\_region | CloudBuild trigger Region. |

## Requirements

### Software

-   [Terraform](https://www.terraform.io/downloads.html) >= 1.3
-   [terraform-provider-google] plugin >= 4.74

### Permissions

- `roles/cloudbuild.builds.editor` on target GCP project
- `roles/storage.objectUser` on supplied GCS folder
- `roles/cloudscheduler.admin` on target GCP project

### APIs

A project with the following APIs enabled must be used to host the
resources of this module and successfully build images:

- Google Compute Engine API: `compute.googleapis.com`
- Google Cloud Build API: `cloudbuild.googleapis.com`
- Google Cloud Scheduler API: `cloudscheduler.googleapis.com`
- Google Identity-Aware Proxy API: `iap.googleapis.com`
- Google Artifact Registry API: `artifactregistry.googleapis.com`
