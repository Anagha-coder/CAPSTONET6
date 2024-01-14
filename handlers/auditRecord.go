package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"example.com/capstone/models"
)

func GenerateAuditRecord(action, itemID string) models.AuditRecord {

	return models.AuditRecord{
		Action:    action,
		ItemID:    itemID,
		Timestamp: time.Now(),
		// PerformedBy: "user123", // You may replace this with the actual user information
	}
}

func PublishAuditRecord(auditRecord models.AuditRecord) error {
	ctx := context.Background()

	// Pub/Sub topic ID
	topicID := "demoTopic"

	// Create a Pub/Sub client
	pubsubClient, err := pubsub.NewClient(ctx, "capstone-408907")
	if err != nil {
		log.Println("Failed to create Pub/Sub client:", err)
		return err
	}

	log.Print("Pub/Sub Client created")

	// Create a Pub/Sub topic client
	topic := pubsubClient.Topic(topicID)

	// Convert audit record to JSON
	auditRecordJSON, err := json.Marshal(auditRecord)
	if err != nil {
		log.Println("Failed to marshal audit record to JSON:", err)
		return err
	}

	// Create a Pub/Sub message
	message := &pubsub.Message{
		Data: auditRecordJSON,
	}

	// Publish the message to the topic
	_, err = topic.Publish(ctx, message).Get(ctx)
	if err != nil {
		log.Println("Failed to publish message to Pub/Sub:", err)
		return err
	}

	log.Println("Audit record published to Pub/Sub successfully")
	return nil
}
