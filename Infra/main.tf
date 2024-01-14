# firestore database
resource "google_firestore_database" "database" {
  count = length(var.databases)

  project                           = var.gcp_project
  name                              = var.databases[count.index].name
  location_id                       = var.gcp_region
  type                              = var.databases[count.index].type
  
}

# Cloud Storage Bucket
resource "google_storage_bucket" "buckets" {
  count = length(var.buckets)
  project                          = var.gcp_project
  name                             = var.buckets[count.index].name
  location                         = var.gcp_region
  uniform_bucket_level_access = true
  
}

resource "google_storage_bucket_object" "objects" {
  count = length(var.objects)

  name   = "${count.index}.zip"
  bucket = google_storage_bucket.buckets[count.index].name
  source = data.archive_file.function_zips[count.index].output_path
}

data "archive_file" "function_zips" {
  count       = length(var.archive_file)
  type        = "zip"
  output_path = "../output/${count.index}.zip"
  source_dir  = "../funcFilesToZip/${count.index}"
}


# cloud functions
resource "google_cloudfunctions2_function" "my_functions" {
  count = length(var.functions)

  name        = var.functions[count.index].name
  location    = var.gcp_region
  description = var.functions[count.index].description

  build_config {
    runtime     = var.functions[count.index].runtime
    entry_point = var.functions[count.index].entry_point

    source {
      storage_source {
        bucket = google_storage_bucket.buckets[count.index].name
        object = google_storage_bucket_object.objects[count.index].name
      }
    }
  }

  event_trigger {
    event_type = var.functions[count.index].event_type
  }

  service_config {
    max_instance_count     = 1
    available_memory       = "256M"
    timeout_seconds        = 60
    service_account_email  = var.service_account_email
  }
}

# resource "google_cloud_run_service_iam_member" "function_members" {
#   count    = length(var.functions)
#   location = google_cloudfunctions2_function.my_functions[count.index].location
#   service  = google_cloudfunctions2_function.my_functions[count.index].name
#   role     = "roles/run.invoker"
#   member   = "allUsers"
# }