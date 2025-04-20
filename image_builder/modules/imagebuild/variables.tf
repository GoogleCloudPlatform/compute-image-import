/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

variable "project_id" {
  type        = string
  description = "Project ID for Image build Pipeline Project."
}

variable "pipeline_name" {
  type        = string
  description = "Name for of the Image build. Cloudbuild trigger will have this name."
}

variable "pipeline_description" {
  type        = string
  description = "Description attached to the cloudbuild trigger."
  default     = ""
}

variable "region" {
  type        = string
  description = "Region used for Cloudbuild trigger."
}

variable "zone" {
  type        = string
  description = "Zone used for cloud build instances."
}

variable "gcs_folder" {
  type        = string
  description = "Folder in which customization scripts are located."
}

variable "source_image" {
  description = <<-EOT
  Initial image to customize, can be one of:
    - image: Full path of image to use.
    - image_family: Full path of image family to use.
    - marketplace_image: Full path of market place image.
    - import: Object describing an image import from a GCS file.
  EOT
  type = object({
    image             = optional(string)
    image_family      = optional(string)
    marketplace_image = optional(string)
    import = optional(object({
      gcs_path           = string # imagebuilder-import-source/vmware_ubuntu22.04_ubuntu22.04-server.vmdk
      license_type       = optional(string)
      generalize         = bool
      skip_os_adaptation = bool
    }))
  })

  validation {
    condition = (
      (var.source_image.image != null ? 1 : 0) +
      (var.source_image.image_family != null ? 1 : 0) +
      (var.source_image.marketplace_image != null ? 1 : 0) +
      (var.source_image.import != null ? 1 : 0)
    ) == 1

    error_message = "Exactly one of image / image_family / marketplace_image / import must be specified."
  }

  validation {
    condition     = var.source_image.image == null ? true : can(regex("projects/(?P<project>[-a-z0-9]+)/global/images/(?P<image>[-a-z0-9]+)", var.source_image.image))
    error_message = "image must be a full path to an image, see selfLink."
  }

  validation {
    condition     = var.source_image.image_family == null ? true : can(regex("projects/(?P<project>[-a-z0-9]+)/global/images/family/(?P<family>[-a-z0-9]+)", var.source_image.image_family))
    error_message = "image_family must be a full path to a image_family."
  }

  validation {
    condition     = var.source_image.marketplace_image == null ? true : can(regex("projects/(?P<project>[-a-z0-9]+)/global/images/(?P<image>[-a-z0-9]+)", var.source_image.marketplace_image))
    error_message = "marketplace_image must be a full path to an image."
  }

  validation {
    condition     = var.source_image.import == null ? true : false
    error_message = "import is currently not supported."
  }
}

variable "target_image_name" {
  type        = string
  description = "Name for target image."
}

variable "target_image_family" {
  type        = string
  description = "Family for target image (optional)."
  default     = ""
}

variable "target_image_description" {
  type        = string
  description = "Description attached to the target image created."
  default     = ""
}

variable "target_image_labels" {
  type        = map(string)
  description = "Key-value pairs of additional labels to attach to the target image."
  default     = {}

  validation {
    condition = alltrue(
      [for key in keys(var.target_image_labels) : can(regex("^[-a-z0-9_]+$", key))]
    )

    error_message = "Only hyphens (-), underscores (_), lowercase characters, and numbers are allowed."
  }

  validation {
    condition = alltrue(
      [for key in values(var.target_image_labels) : can(regex("^[-a-z0-9_]+$", key))]
    )

    error_message = "Only hyphens (-), underscores (_), lowercase characters, and numbers are allowed."
  }
}

variable "target_image_region" {
  type        = string
  description = "Region used for target image."
  default     = ""
}

variable "target_image_encryption_key" {
  type        = string
  description = "CMEK used for target image (and processing)."
  default     = ""

  validation {
    condition = var.target_image_encryption_key == "" ? true : can(regex("^projects/[^/]+/locations/global/keyRings/[^/]+/cryptoKeys/[^/]+$", var.target_image_encryption_key))

    error_message = "Should be a valid path to a key ring."
  }
}

variable "customization_script_source" {
  type        = string
  description = "Path to local customize script."
  default     = ""
}

variable "testing_script_source" {
  type        = string
  description = "Path to local test script."
  default     = ""
}

variable "schedule_cron_pattern" {
  type        = string
  description = "Optional cron pattern for setting up a schedule."
  default     = ""
}

variable "schedule_timezone" {
  type        = string
  description = "Optional timezone for the schedule. Must be a time zone name from the tz database."
  default     = "UTC"
}

variable "service_account_id" {
  type        = string
  description = "ID of the service account to use."
}
