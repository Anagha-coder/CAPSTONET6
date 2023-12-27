package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"google.golang.org/api/iterator"
)

// Q: Create a new grocery item
// Any new item created would be stored into the data source
// The image for the item should be uploaded to a storage bucket - connection with cloud storage bucket!!
// A thumbnail for the image should be generated and stored on the bucket (asynchronously)

func CreateGroceryItem(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight request
	if r.Method == http.MethodOptions {
		// Set CORS headers for preflight requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	utils.InitLogger()

	// Log the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read request body:", err)
	}
	log.Println("Request Body:", string(body))
	log.Print("Decode started:")

	// schema reference based on that to create new item ✅
	var groceryItem models.GroceryItem
	if err := json.Unmarshal(body, &groceryItem); err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	// conn firestore database to store the grocery items ✅
	// Create a Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create Firestore client")
		return
	}
	defer client.Close()

	log.Print("Firestore client created")

	// Read existing grocery items from Firestore (assuming you have a collection named "groceryItems")
	iter := client.Collection("groceryItems").Documents(context.Background())
	var existingGroceryItems []models.GroceryItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Print("Failed to read grocery item data from Firestore:", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to read grocery item data from Firestore")
			return
		}
		var item models.GroceryItem
		if err := doc.DataTo(&item); err != nil {
			log.Print("Failed to parse grocery item data from Firestore:", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to parse grocery item data from Firestore")
			return
		}
		existingGroceryItems = append(existingGroceryItems, item)
	}

	log.Print("Existing grocery items read from Firestore")

	// Generate a unique ID for the new grocery item
	newItemID := generateUniqueGroceryItemID(existingGroceryItems)

	// Set the new grocery item ID
	groceryItem.ID = newItemID

	// Add the new grocery item to Firestore
	_, _, err = client.Collection("groceryItems").Add(context.Background(), groceryItem)
	if err != nil {
		log.Print("Failed to create grocery item in Firestore:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create grocery item in Firestore")
		return
	}

	log.Print("Grocery item created successfully in Firestore")

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "Grocery item created successfully"})
	log.Print("Response Sent: CreateGroceryItem")

}

func generateUniqueGroceryItemID(existingGroceryItems []models.GroceryItem) int {
	highestID := 0

	// Find the highest existing ID from the Firestore data
	for _, item := range existingGroceryItems {
		if item.ID > highestID {
			highestID = item.ID
		}
	}

	// Increment the highest existing ID to generate a new unique ID
	return highestID + 1
}

//cloud storage bucket connection - to store images

// how thumbnails will be stored?
