package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("FetchItemByID", FetchItemByID)
}

// FetchItemByID fetches a grocery item by its ID.
// @Summary Fetch a grocery item by ID
// @Description Fetches a grocery item from the Firestore database based on the provided ID.
// @ID fetch-item-by-id
// @Produce json
// @Param id path integer true "ID of the grocery item" format(int64) minimum(1)
// @Success 200 {object} GroceryItem "Grocery item fetched successfully"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 404 {object} ErrorResponse "Not Found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /fetchGroceryItemByID/{id} [get]
func FetchItemByID(w http.ResponseWriter, r *http.Request) {
	// handle preflight CORS
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Contet-Type")
		w.WriteHeader(http.StatusOK)
	}

	utils.InitLogger()

	uri := r.RequestURI

	// Split url in parts "/"
	parts := strings.Split(uri, "/")

	// check if URI has 2 parts and last part is valid int
	if len(parts) < 2 {
		log.Println("Invalid URL format")
		respondWithError(w, http.StatusBadRequest, "Invalid URI format")
		return
	}

	// fetch productId from url
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		log.Print("Requested Id is invalid", err)
		respondWithError(w, http.StatusBadRequest, "Request Id is invalid")
	}

	log.Print("Request received: FetchItemByID, ID:", id)

	// connection to firestore
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client ", err)
		respondWithError(w, http.StatusBadRequest, "failed to create firestore client")
		return
	}
	defer client.Close()

	// collection - groceryItem- query by id - return info & img
	query := client.Collection("groceryItems").Where("ID", "==", id)

	iter := query.Documents(context.Background())
	doc, err := iter.Next()
	if err != nil {
		log.Print("GroceryItem not found:", err)
		respondWithError(w, http.StatusBadRequest, "GroceryItem not found, maybe it does not exist")
		return
	}

	var groceryItem models.GroceryItem
	if err := doc.DataTo(&groceryItem); err != nil {
		log.Print("Failed to parse data: ", err)
		respondWithError(w, http.StatusBadRequest, "Failed to parse data")
		return
	}

	log.Print("Sending response: FetchItemByID")
	respondWithJSON(w, http.StatusOK, groceryItem)

}

// may be go routine to return image and thumbnail - async - no need
