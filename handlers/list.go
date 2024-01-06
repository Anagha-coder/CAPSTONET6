package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
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

	// Create a Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create Firestore client")
		return
	}
	defer client.Close()

	log.Print("Firestore client created")

	collectionRef := client.Collection("groceryItems")
	var query firestore.Query

	i := 0

	for k := range r.URL.Query() {
		v := r.URL.Query().Get(k)

		// Handle value inequality and range comparisons
		if strings.HasPrefix(k, "price") {
			// Convert the value to a float64 for numerical comparison
			price, err := strconv.ParseFloat(v, 64)
			if err != nil {
				log.Printf("Invalid price value: %s", v)
				continue
			}

			if strings.HasSuffix(k, "_min") {
				query = collectionRef.Where("Price", "<=", price)
			} else if strings.HasSuffix(k, "_max") {
				query = collectionRef.Where("Price", ">=", price)
			} else {
				query = collectionRef.Where("Price", "==", price)
			}

		} else {
			if i == 0 {
				log.Printf("Query parameter: %s=%s\n", k, v)
				query = collectionRef.Where(k, "==", v)

			} else {
				query = query.Where(k, "==", v)
			}

			i++

		}

	}

	// try to add "productname" containing baby care oil not exact name - but keyword

	log.Printf("Request Parameters: %v", r.URL.Query())

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
