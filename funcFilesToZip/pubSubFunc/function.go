package helloworld

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

// MessagePublishedData contains the full Pub/Sub message
type MessagePublishedData struct {
	Message PubSubMessage `json:"message"`
}

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

func init() {
	functions.CloudEvent("HelloPubSub", helloPubSub)
}

func helloPubSub(ctx context.Context, e event.Event) error {
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}

	name := string(msg.Message.Data) // Automatically decoded from base64.
	if name == "" {
		name = "World"
	}
	log.Printf("Hello, %s!", name)

	// Store the message in Firestore
	if err := storeInFirestore(ctx, name); err != nil {
		return fmt.Errorf("failed to store in Firestore: %v", err)
	}

	return nil
}

func storeInFirestore(ctx context.Context, message string) error {
	// Set up the Firestore client
	client, err := firestore.NewClient(ctx, "capstone-408907")
	if err != nil {
		return fmt.Errorf("firestore.NewClient: %v", err)
	}
	defer client.Close()

	// Reference to the Firestore collection
	collection := client.Collection("audit-records")

	// Create a document with the message
	_, _, err = collection.Add(ctx, map[string]interface{}{
		"message": message,
	})
	if err != nil {
		return fmt.Errorf("collection.Add: %v", err)
	}

	log.Printf("Message stored in Firestore: %s", message)
	return nil
}
