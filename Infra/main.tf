# firestore database
resource "google_firestore_database" "database" {
  count = length(var.databases)

  project                           = var.gcp_project
  name                              = var.databases[count.index].name
  location_id                       = var.gcp_region
  type                              = var.databases[count.index].type
  
}

resource "google_storage_bucket" "buckets" {
  count = length(var.buckets)
  project                          = var.gcp_project
  name                             = var.buckets[count.index].name
  location                         = var.gcp_region
  uniform_bucket_level_access = true
  
}


