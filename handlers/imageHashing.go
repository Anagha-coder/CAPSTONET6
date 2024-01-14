package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"mime/multipart"
	"strings"

	"example.com/capstone/utils"
	"google.golang.org/api/iterator"
)

func CalculateImageHash(file multipart.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

func IsDuplicateImage(imageHash string) bool {
	// Create a Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		// respondWithError(w, http.StatusInternalServerError, "Failed to create Firestore client")
		return false
	}
	defer client.Close()

	log.Print("Firestore client created to check ImageHash")

	log.Println("Checking for duplicate image with hash:", imageHash)

	cleanedHash := strings.Trim(imageHash, "\"")

	// Query Firestore for documents with the given imageHash
	iter := client.Collection("groceryItems").Where("ImageHash", "==", cleanedHash).Documents(context.Background())

	// Flag to track duplicate status
	duplicateFound := false

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			// No more documents to check
			break
		}
		if err != nil {
			log.Print("Failed to check duplicate imageHash:", err)
			return false
		}
		// Debug: Print document ID and data
		log.Printf("Found document ID: %s", doc.Ref.ID)
		log.Printf("Document data: %+v", doc.Data())

		// Set flag to true since duplicate image found
		duplicateFound = true

	}

	log.Println("Duplicate found:", duplicateFound)

	// Return the final duplicate status
	return duplicateFound
}
