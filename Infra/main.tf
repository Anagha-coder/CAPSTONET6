# firestore database
# resource "google_firestore_database" "database" {
#   count = length(var.databases)

#   project                           = var.gcp_project
#   name                              = var.databases[count.index].name
#   location_id                       = var.gcp_region
#   type                              = var.databases[count.index].type
  
# }

# Cloud Storage Bucket
resource "google_storage_bucket" "buckets" {
  count = length(var.buckets)
  project                          = var.gcp_project
  name                             = var.buckets[count.index].name
  location                         = var.gcp_region
  uniform_bucket_level_access = true
  
}

# Bucket to store cloud functions
resource "google_storage_bucket" "bucket" {
    name = "cloud-function-bucket-by-anagha-capstone"
    location = "US"
    uniform_bucket_level_access = true
  
}

# Cloud functions
#1
resource "google_storage_bucket_object" "object" {
  name = "createGrocery.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.createGrocery_zip.output_path
}

data "archive_file" "createGrocery_zip" {
  type        = "zip"
  output_path = "../output/createGrocery.zip"
  source_dir  = "../funcFilesToZip/createCAP"
}

resource "google_cloudfunctions2_function" "createGroceryItem" {
  name = "createGroceryItem"
  location = var.gcp_region
  description = "Create a new Grocery Item"
  build_config {
    runtime = "go121"
    entry_point ="CreateGroceryItem"

    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.object.name

      }
    }
    
  }

  service_config {
    max_instance_count = 1
    available_memory = "256M"
    timeout_seconds = 60
    service_account_email = var.service_account_email
  }
  
}

resource "google_cloud_run_service_iam_member" "member" {
  location = google_cloudfunctions2_function.createGroceryItem.location
  service  = "creategroceryitem"
  role     = "roles/run.invoker"
  member   = "allUsers"
}


# cf2
resource "google_storage_bucket_object" "object2" {
  name = "updateCAP.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.updateCAP_zip.output_path
}

data "archive_file" "updateCAP_zip" {
  type        = "zip"
  output_path = "../output/updateCAP.zip"
  source_dir  = "../funcFilesToZip/updateCAP"
}

resource "google_cloudfunctions2_function" "updateCAP" {
  name = "updateGroceryItemByID"
  location = var.gcp_region
  description = "Update an existing Grocery Item"
  

  build_config {
    runtime = "go121"
    entry_point ="UpdateGroceryItem"

    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.object2.name

      }
    }
    
  }

  service_config {
    max_instance_count = 1
    available_memory = "256M"
    timeout_seconds = 60
    service_account_email = var.service_account_email
    
    
  }  
}

resource "google_cloud_run_service_iam_member" "member2" {
  location = google_cloudfunctions2_function.updateCAP.location
  service  = "updategroceryitembyid"
  role     = "roles/run.invoker"
  member   = "allUsers"
}


# cf3 
resource "google_storage_bucket_object" "object3" {
  name = "deleteCAP.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.deleteCAP_zip.output_path
}

data "archive_file" "deleteCAP_zip" {
  type        = "zip"
  output_path = "../output/deleteCAP.zip"
  source_dir  = "../funcFilesToZip/deleteCAP"
}

resource "google_cloudfunctions2_function" "deleteCAP" {
  name = "deleteGroceryItemByID"
  location = var.gcp_region
  description = "Deletes a grocery Item"

  build_config {
    runtime = "go121"
    entry_point ="DeleteItemByID"

    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.object3.name

      }
    }
    
  }

  service_config {
    max_instance_count = 1
    available_memory = "256M"
    timeout_seconds = 60
    service_account_email = var.service_account_email
    
  }
  
}

resource "google_cloud_run_service_iam_member" "member3" {
  location = google_cloudfunctions2_function.deleteCAP.location
  service  = "deletegroceryitembyid"
  role     = "roles/run.invoker"
  member   = "allUsers"
}



# cf4
resource "google_storage_bucket_object" "object4" {
  name = "fetchCAP.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.fetchCAP_zip.output_path
}

data "archive_file" "fetchCAP_zip" {
  type        = "zip"
  output_path = "../output/fetchCAP.zip"
  source_dir  = "../funcFilesToZip/fetchCAP"
}

resource "google_cloudfunctions2_function" "fetchCAP" {
  name = "fetchGroceryItemByID"
  location = var.gcp_region
  description = "Fetch Grocery Item By ID"

  build_config {
    runtime = "go121"
    entry_point ="FetchItemByID"

    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.object4.name

      }
    }
    
  }

  service_config {
    max_instance_count = 1
    available_memory = "256M"
    timeout_seconds = 60
    service_account_email = var.service_account_email
    
  }
  
}

resource "google_cloud_run_service_iam_member" "member4" {
  location = google_cloudfunctions2_function.fetchCAP.location
  service  = "fetchgroceryitembyid"
  role     = "roles/run.invoker"
  member   = "allUsers"
}



# cf5
resource "google_storage_bucket_object" "object5" {
  name = "listCAP.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.listCAP_zip.output_path
}

data "archive_file" "listCAP_zip" {
  type        = "zip"
  output_path = "../output/listCAP.zip"
  source_dir  = "../funcFilesToZip/listCAP"
}

resource "google_cloudfunctions2_function" "listCAP" {
  name = "listGroceryItems"
  location = var.gcp_region
  description = "List Grocery Items By categories"

  build_config {
    runtime = "go121"
    entry_point ="ListItemsBY"

    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.object5.name

      }
    }
    
  }

  service_config {
    max_instance_count = 1
    available_memory = "256M"
    timeout_seconds = 60
    service_account_email = var.service_account_email
    
  }
  
}

resource "google_cloud_run_service_iam_member" "member5" {
  location = google_cloudfunctions2_function.listCAP.location
  service  = "listgroceryitems"
  role     = "roles/run.invoker"
  member   = "allUsers"
}

#cf6
resource "google_storage_bucket_object" "object6" {
  name = "bulkCAP.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.bulkCAP_zip.output_path
}

data "archive_file" "bulkCAP_zip" {
  type        = "zip"
  output_path = "../output/bulkCAP.zip"
  source_dir  = "../funcFilesToZip/bulkFuncCAP"
}

resource "google_cloudfunctions2_function" "bulkCAP" {
  name = "BulkUploadGroceryItems"
  location = var.gcp_region
  description = "Bulk Uploads Grocery Items"

  build_config {
    runtime = "go121"
    entry_point ="BulkUpload"

    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.object6.name

      }
    }
    
  }

  service_config {
    max_instance_count = 1
    available_memory = "256M"
    timeout_seconds = 60
    service_account_email = var.service_account_email
    
  }
  
}

resource "google_cloud_run_service_iam_member" "member6" {
  location = google_cloudfunctions2_function.bulkCAP.location
  service  = "bulkuploadgroceryitems"
  role     = "roles/run.invoker"
  member   = "allUsers"
}