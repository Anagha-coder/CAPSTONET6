package handlers

import (
	"context"
	"encoding/json"

	"fmt"
	"log"
	"net/http"

	"strconv"
	"strings"
	"time"

	"example.com/capstone/utils"
	"google.golang.org/api/iterator"
)

func UpdateGroceryItem(w http.ResponseWriter, r *http.Request) {
	// handle preflight CORS
	if r.Method == http.MethodOptions {
		w.Header().Set("Access Control Allow-Origin ", "*")
		w.Header().Set("Access Control Allow-Methods", "UPDATE, OPTIONS")
		w.Header().Set("Access Control Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
	}

	utils.InitLogger()
	log.Print("Request is being Processed for Updating exsisting GroceryItem")

	uri := r.RequestURI

	parts := strings.Split(uri, "/")

	// checks uri is valid or not
	if len(parts) < 2 {
		log.Print("Invalid Request URI")
		respondWithError(w, http.StatusBadRequest, "Invalid Request URI")
		return
	}

	// Extract productid from uri
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		log.Fatalf("Unable to parse %q as int: %v", parts[len(parts)-1], err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	log.Print("The request was made for ID:", id)

	// the request will be a multipart form json- data and image
	// Parse the form data with a max of 10 MB limit for the entire request
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println("Failed to parse multipart form:", err)
		respondWithError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get the JSON data from the form
	jsonData := r.FormValue("json-data")
	if jsonData == "" {
		log.Print("JSON data is required to create grocery item.")
		respondWithError(w, http.StatusBadRequest, "No 'json-data' field provided in the form")
	}

	// schema reference based on that to create new item
	var updatedGroceryItem map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &updatedGroceryItem); err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if err := validateRequiredFields(updatedGroceryItem); err != nil {
		log.Println("Missing required fields:", err)
		respondWithError(w, http.StatusBadRequest, "Missing required fields: "+err.Error())
		return
	}

	// Get the image file from the form data
	file, _, err := r.FormFile("image")
	// log.Printf("Original image format: %s", formatimg)
	if err == http.ErrMissingFile {
		// no image provided, proceed without image
		log.Println("No image file")
	} else if err != nil {
		log.Println("Failed to get image file:", err)
		respondWithError(w, http.StatusBadRequest, "Failed to get image file")
		return
	} else {
		defer file.Close()

		// Use a channel to signal when the image upload is complete
		imageUploadDone := make(chan bool)

		// a goroutine to handle image upload asynchronously
		go func() {
			// Upload the image to the Cloud Storage bucket
			imageURL, thumbnailURL, err := ImageAndThumbnailUploadFunc(file, updatedGroceryItem)
			if err != nil {
				log.Println("failed to upload thumbnail to cloud storage:", err)
				// Handle error if needed
				imageUploadDone <- false
				return
			}

			// Set the image URL in the grocery item
			updatedGroceryItem["Image"] = imageURL
			updatedGroceryItem["Thumbnail"] = thumbnailURL

			// Signal that the image upload is complete
			imageUploadDone <- true
		}()

		// Wait for the image upload to complete (or for a timeout)
		select {
		case <-imageUploadDone:
			log.Println("Image uploaded!")
		case <-time.After(30 * time.Second): // Set a timeout if needed
			log.Println("Image upload timed out")
			// Handle timeout if needed
		}

	}

	// connection with firestore
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("failed to create a firestore connetion")
		respondWithError(w, http.StatusBadRequest, " failed to create a firestore collection")
		return
	}
	defer client.Close()

	// query firestore to check if the item with the given ID exists
	query := client.Collection("groceryItems").Where("ID", "==", id).Limit(1)
	iter := query.Documents(context.Background())
	doc, err := iter.Next()

	if err == iterator.Done {
		log.Print("Grocery item not found")
		respondWithError(w, http.StatusNotFound, "Grocery item not found")
		return
	} else if err != nil {
		log.Print("Failed to read grocery item data from Firestore:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to read grocery item data from Firestore")
		return
	}

	docRef := client.Collection("groceryItems").Doc(doc.Ref.ID)

	// Set the existing ID to updatedGroceryItem
	updatedGroceryItem["ID"] = id

	// Unmarshal the JSON data into the existing grocery item
	if err := json.Unmarshal([]byte(jsonData), &updatedGroceryItem); err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Update existing fields with new values
	_, err = docRef.Set(context.Background(), updatedGroceryItem)
	if err != nil {
		log.Print("Failed to update grocery item in Firestore:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update grocery item in Firestore")
		return
	}

	// Generate audit record for update
	auditRecord := GenerateAuditRecord("update", strconv.Itoa(id))

	// Print audit record to log
	log.Printf("Audit Record: %+v", auditRecord)

	// Publish audit record to Pub/Sub topic
	some_err := PublishAuditRecord(auditRecord)
	if some_err != nil {

		log.Println("Failed to publish audit record:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to publish audit record")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Grocery item updated successfully"})
	log.Print("Response Sent: UpdateGroceryItem")

}

func validateRequiredFields(item map[string]interface{}) error {
	// Check required fields
	requiredFields := []string{"productName", "category", "price", "weight", "weightUnit", "manufacturer", "brand", "itemPackageQuantity", "packageInformation", "mfgDate", "expDate", "countryOfOrigin"}

	for _, field := range requiredFields {
		if value, ok := item[field]; !ok || isEmpty(value) {
			return fmt.Errorf("field '%s' is required", field)
		}
	}

	return nil
}

// function to check if a value is considered empty
func isEmpty(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0.0
	case bool:
		return !v
	case nil:
		return true
	default:
		return false
	}
}
