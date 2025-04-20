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

output "trigger_id" {
  description = "CloudBuild trigger ID."
  value       = google_cloudbuild_trigger.image_build_trigger.id
}

output "trigger_name" {
  description = "CloudBuild trigger Name."
  value       = google_cloudbuild_trigger.image_build_trigger.name
}

output "trigger_region" {
  description = "CloudBuild trigger Region."
  value       = google_cloudbuild_trigger.image_build_trigger.location
}
