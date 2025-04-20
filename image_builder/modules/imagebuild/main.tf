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

resource "google_storage_bucket_object" "customize_script" {
  count  = var.customization_script_source == "" ? 0 : 1
  name   = "${local.gcs_prefix}/customize.pkr.hcl"
  source = var.customization_script_source
  bucket = local.gcs_bucket
}

resource "google_storage_bucket_object" "test_script" {
  count  = var.testing_script_source == "" ? 0 : 1
  name   = "${local.gcs_prefix}/test.pkr.hcl"
  source = var.testing_script_source
  bucket = local.gcs_bucket
}

resource "google_cloud_scheduler_job" "trigger_schedule" {
  count = var.schedule_cron_pattern == "" ? 0 : 1

  name      = var.pipeline_name
  project   = var.project_id
  region    = var.region
  schedule  = var.schedule_cron_pattern
  time_zone = var.schedule_timezone

  http_target {
    http_method = "POST"
    uri         = "https://cloudbuild.googleapis.com/v1/projects/${var.project_id}/locations/${var.region}/triggers/${var.pipeline_name}:run"
    body = base64encode(
      <<-EOT
        {
          projectId: '${var.project_id}',
          triggerId: '${var.pipeline_name}',
        }
        EOT
    )
    oauth_token {
      scope                 = "https://www.googleapis.com/auth/cloud-platform"
      service_account_email = local.sa_email
    }
  }
}

resource "google_cloudbuild_trigger" "image_build_trigger" {
  project     = var.project_id
  name        = var.pipeline_name
  description = var.pipeline_description
  location    = var.region
  substitutions = merge(
    {
      _PROJECT_ID                  = var.project_id
      _ZONE                        = var.zone
      _GCS_FOLDER                  = var.gcs_folder
      _TARGET_IMAGE_NAME           = var.target_image_name
      _TARGET_IMAGE_FAMILY         = var.target_image_family
      _TARGET_IMAGE_DESCRIPTION    = var.target_image_description
      _TARGET_IMAGE_REGION         = var.target_image_region == "" ? var.region : var.target_image_region
      _TARGET_IMAGE_ENCRYPTION_KEY = var.target_image_encryption_key
      _TARGET_IMAGE_LABELS         = jsonencode(var.target_image_labels)
      _TARGET_IMAGE_LABELS_FLAT    = join(",", [for k, v in var.target_image_labels : "${k}=${v}"])
  }, local.src_subs)
  repository_event_config {}

  lifecycle {
    ignore_changes = [
      repository_event_config
    ]
  }
  service_account = var.service_account_id

  tags = ["imagebuilder"]

  build {
    dynamic "step" {
      for_each = flatten([
        local.first_step,
        var.customization_script_source == "" ? [local.create_target_image_step] : [],
        var.customization_script_source == "" && var.testing_script_source == "" ? [] : [local.copy_packer_scripts_step],
        var.customization_script_source != "" ? [local.customization_init_step, local.customization_step] : [],
        local.add_label_step,
        local.delete_temp_source_image_step,
        var.testing_script_source != "" ? [local.test_init_step, local.test_step] : [],
        var.target_image_family != "" ? [local.set_target_family_step] : []
      ])
      iterator = custom_step
      content {
        name = custom_step.value.name
        env  = custom_step.value.env
        args = custom_step.value.args
        id   = custom_step.value.id
      }
    }

    options {
      logging = "CLOUD_LOGGING_ONLY"
    }

    substitutions = {
      _SUFFIXED_TARGET_IMAGE_NAME = "$${_TARGET_IMAGE_NAME}-$${BUILD_ID:0:8}"
      _SUFFIXED_SOURCE_IMAGE_NAME = "$${_TARGET_IMAGE_NAME}-source-$${BUILD_ID:0:8}"
    }

    timeout = "3600s"
  }
}

locals {
  image_import_license = "https://www.googleapis.com/compute/v1/projects/prod-vmmig-images-public/global/licenses/image-builder-tf-license"

  packer_image = "${var.region}-docker.pkg.dev/${var.project_id}/packer/packer"

  gcs_bucket = split("/", trimprefix(var.gcs_folder, "gs://"))[0]
  gcs_prefix = trim(trimprefix(trimprefix(var.gcs_folder, "gs://"), local.gcs_bucket), "/")

  sa_split = split("/", var.service_account_id)
  sa_email = element(local.sa_split, length(local.sa_split) - 1)
}

locals {
  src_image        = var.source_image.image != null ? regex("projects/(?P<project>[-a-z0-9]+)/global/images/(?P<image>[-a-z0-9]+)", var.source_image.image) : null
  src_family       = var.source_image.image_family != null ? regex("projects/(?P<project>[-a-z0-9]+)/global/images/family/(?P<family>[-a-z0-9]+)", var.source_image.image_family) : null
  src_market_place = var.source_image.marketplace_image != null ? regex("projects/(?P<project>[-a-z0-9]+)/global/images/(?P<image>[-a-z0-9]+)", var.source_image.marketplace_image) : null

  src_subs = (
    local.src_image != null ? (
      {
        _SOURCE_IMAGE    = local.src_image.image
        _SOURCE_IMAGE_OS = local.src_image.project
      }
      ) : local.src_family != null ? (
      {
        _SOURCE_IMAGE_FAMILY    = local.src_family.family
        _SOURCE_IMAGE_FAMILY_OS = local.src_family.project
      }
      ) : local.src_market_place != null ? (
      {
        _MARKETPLACE_IMAGE         = local.src_market_place.image
        _MARKETPLACE_IMAGE_PROJECT = local.src_market_place.project
      }

      ) : var.source_image.import != null ? (
      {
        _IMPORT_GCS_FILE            = "TBD"
        _IMPORT_SKIP_OS_ADAPTATION  = "TBD"
        _IMPORT_GENERALIZE          = "TBD"
        _IMPORT_LICENSE_TYPE        = "TBD"
        _IMPORT_ADDITIONAL_LICENSES = "TBD"
        _IMPORT_OS                  = "TBD"
        _IMPORT_CPU_ARCHITECTURE    = "TBD"
        _IMPORT_LOCATION            = "TBD"
        _IMPORT_TARGET_PROJECT      = "TBD"
      }
    ) : {}
  )
}

locals {
  create_temporary_source_image_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "compute", "images", "create", "$${_SUFFIXED_SOURCE_IMAGE_NAME}",
      "--source-image=$${_SOURCE_IMAGE}",
      "--source-image-project=$${_SOURCE_IMAGE_OS}",
      "--licenses=${local.image_import_license}"
    ]
    id = "create-temporary-source-image"
  }
  create_temporary_source_image_from_family_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "compute", "images", "create", "$${_SUFFIXED_SOURCE_IMAGE_NAME}",
      "--source-image-family=$${_SOURCE_IMAGE_FAMILY}",
      "--source-image-project=$${_SOURCE_IMAGE_FAMILY_OS}",
      "--licenses=${local.image_import_license}"
    ]
    id = "create-temporary-source-image"
  }
  copy_market_place_image_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "compute", "images", "create", "$${_SUFFIXED_SOURCE_IMAGE_NAME}",
      "--source-image=$${_MARKETPLACE_IMAGE}",
      "--source-image-project=$${_MARKETPLACE_IMAGE_PROJECT}",
      "--licenses=${local.image_import_license}"
    ]
    id = "copy-marketplace-image"
  }
  import_image_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    # entrypoint = "bash"
    args = ["-eEuo", "pipefail", "-c",
      # TBD
    ]
    id = "import-image"
  }

  first_step = (
    local.src_image != null ? (
      local.create_temporary_source_image_step
      ) : local.src_family != null ? (
      local.create_temporary_source_image_from_family_step
      ) : local.src_market_place != null ? (
      local.copy_market_place_image_step
    ) : local.import_image_step
  )

  create_target_image_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "compute", "images", "create", "$${_SUFFIXED_TARGET_IMAGE_NAME}",
      "--description=$${_TARGET_IMAGE_DESCRIPTION}",
      "--storage-location=$${_TARGET_IMAGE_REGION}",
      "--kms-key=$${_TARGET_IMAGE_ENCRYPTION_KEY}",
      "--labels=$${_TARGET_IMAGE_LABELS_FLAT}",
      "--source-image=$${_SUFFIXED_SOURCE_IMAGE_NAME}"
    ]
    id = "create-target-image"
  }
  copy_packer_scripts_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "storage", "cp", "gs://$${_GCS_FOLDER}*.pkr.hcl", "."]
    id   = "copy-packer-scripts"
  }
  customization_init_step = {
    name = local.packer_image
    env  = []
    args = ["init", "customize.pkr.hcl"]
    id   = "initialize-customization"
  }
  customization_step = {
    name = local.packer_image
    env = [
      "PKR_VAR_project_id=$${_PROJECT_ID}",
      "PKR_VAR_zone=$${_ZONE}",
      "PKR_VAR_target_image_name=$${_SUFFIXED_TARGET_IMAGE_NAME}",
      "PKR_VAR_target_image_description=$${_TARGET_IMAGE_DESCRIPTION}",
      "PKR_VAR_target_image_region=$${_TARGET_IMAGE_REGION}",
      "PKR_VAR_target_image_encryption_key=$${_TARGET_IMAGE_ENCRYPTION_KEY}",
      "PKR_VAR_target_image_labels=$${_TARGET_IMAGE_LABELS}",
      "PKR_VAR_source_image=$${_SUFFIXED_SOURCE_IMAGE_NAME}",
    ]
    args = ["build", "customize.pkr.hcl"]
    id   = "customize-image"
  }
  add_label_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "compute", "images", "add-labels", "$${_SUFFIXED_TARGET_IMAGE_NAME}", "--labels=pipeline=$${TRIGGER_NAME},build-id=$${BUILD_ID}"]
    id   = "add-pipeline-label"
  }
  delete_temp_source_image_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "compute", "images", "delete", "$${_SUFFIXED_SOURCE_IMAGE_NAME}"]
    id   = "delete-temporary-source-image"
  }
  test_init_step = {
    name = local.packer_image
    env  = []
    args = ["init", "test.pkr.hcl"]
    id   = "initialize-test"
  }
  test_step = {
    name = local.packer_image
    env = [
      "PKR_VAR_project_id=$${_PROJECT_ID}",
      "PKR_VAR_zone=$${_ZONE}",
      "PKR_VAR_target_image_name=$${_SUFFIXED_TARGET_IMAGE_NAME}",
      "PKR_VAR_target_image_encryption_key=$${_TARGET_IMAGE_ENCRYPTION_KEY}",
    ]
    args = ["build", "test.pkr.hcl"]
    id   = "test-image"
  }
  set_target_family_step = {
    name = "gcr.io/google.com/cloudsdktool/cloud-sdk"
    env  = []
    args = ["gcloud", "compute", "images", "update", "$${_SUFFIXED_TARGET_IMAGE_NAME}", "--family=$${_TARGET_IMAGE_FAMILY}"]
    id   = "set-image-family"
  }
}
