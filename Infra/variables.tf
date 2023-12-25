variable "gcp_svc_key" {
  description = "GCP service account key"
}

variable "gcp_project" {
  description = "GCP project ID"
}

variable "gcp_region" {
    description = "GCP region"  
}

variable "databases" {
  description = "List of databases to create"
  type        = list(object({
    name         = string
    location  = string
    type         = string
  }))
}

variable "buckets" {
  description = "Bucket to store images of grocery items"
  type = list(object({
    name = string
    location = string
  }))
  
}