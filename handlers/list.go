package handlers

import (
	"context"
	"log"
	"net/http"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"google.golang.org/api/iterator"
)

func ListItemsBY(w http.ResponseWriter, r *http.Request) {
	// handle preflight CORS
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		w.WriteHeader(http.StatusOK)
	}

	utils.InitLogger()

	// Parse query parameters for filters
	productNameFilter := r.URL.Query().Get("productName")
	// categoryFilter := r.URL.Query().Get("category")
	// priceFilter := r.URL.Query().Get("price")

	// Create a Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create Firestore client")
		return
	}
	defer client.Close()

	log.Print("Firestore client created")

	query := client.Collection("groceryItems").Where("ProductName", "==", productNameFilter)

	// Execute the query
	iter := query.Documents(context.Background())

	var groceryItems []models.GroceryItem
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
		groceryItems = append(groceryItems, item)
	}

	// Create a response object
	var response []interface{}
	for _, item := range groceryItems {
		itemMap := map[string]interface{}{
			"ID":          item.ID,
			"ProductName": item.ProductName,
			"Price":       item.Price,
			"Category":    item.Category,
			"Thumbnail":   item.Thumbnail,
		}
		response = append(response, itemMap)
	}

	// Return the response as JSON
	respondWithJSON(w, http.StatusOK, response)
	log.Print("Response Sent: ListGroceryItems")
}
