// Package helloworld provides a set of Cloud Functions samples.
package helloworld

import (
  "context"
  "fmt"
  "log"
  "time"

  "github.com/GoogleCloudPlatform/functions-framework-go/functions"
  "github.com/cloudevents/sdk-go/v2/event"
)

func init() {
  functions.CloudEvent("HelloStorage", helloStorage)
}

// StorageObjectData contains metadata of the Cloud Storage object.
type StorageObjectData struct {
  Bucket         string    `json:"bucket,omitempty"`
  Name           string    `json:"name,omitempty"`
  Metageneration int64     `json:"metageneration,string,omitempty"`
  TimeCreated    time.Time `json:"timeCreated,omitempty"`
  Updated        time.Time `json:"updated,omitempty"`
}

// helloStorage consumes a CloudEvent message and logs details about the changed object.
func helloStorage(ctx context.Context, e event.Event) error {
  log.Printf("Event ID: %s", e.ID())
  log.Printf("Event Type: %s", e.Type())

  var data StorageObjectData
  if err := e.DataAs(&data); err != nil {
    return fmt.Errorf("event.DataAs: %v", err)
  }

  log.Printf("Bucket: %s", data.Bucket)
  log.Printf("File: %s", data.Name)
  log.Printf("Metageneration: %d", data.Metageneration)
  log.Printf("Created: %s", data.TimeCreated)
  log.Printf("Updated: %s", data.Updated)
  return nil
}
