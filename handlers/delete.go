package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"example.com/capstone/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DeleteItemByID(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight request
	if r.Method == http.MethodOptions {
		// Set CORS headers for preflight requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	utils.InitLogger()

	log.Print("Request is being Processed for Deleting Item by ID")

	uri := r.RequestURI

	// Split the URI using '/'
	parts := strings.Split(uri, "/")

	// Check if the URI has at least 4 parts and the last part is a valid integer
	if len(parts) < 2 {
		log.Print("Invalid URI format")
		respondWithError(w, http.StatusBadRequest, "Invalid URI format")
		return
	}

	// Extract the last part of the URL as the employee ID
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		log.Print("Invalid Item ID:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid Item ID")
		return
	}

	log.Print("Request received: DeleteItem, ID:", id)

	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create Firestore client")
		return
	}
	defer client.Close()

	log.Print("Request received: DeleteItem by ID")

	// Define a query to retrieve the document with the specified "ID" field value
	query := client.Collection("groceryItems").Where("ID", "==", id).Limit(1)

	// Run the query
	iter := query.Documents(context.Background())
	doc, err := iter.Next()

	if err != nil {
		log.Print("Failed to retrieve item from Firestore:", err)
		if status.Code(err) == codes.NotFound {
			respondWithError(w, http.StatusNotFound, "Item not found")
		} else {
			// Handle other errors
			respondWithError(w, http.StatusInternalServerError, "Item not found. Check requested ID")
		}
		return
	}

	_, err = doc.Ref.Delete(context.Background())
	if err != nil {
		log.Print("failed to delete item from Firestore database:", err)
		respondWithError(w, http.StatusInternalServerError, "failed to delete item from Firestore database")
		return
	}

	log.Print("Item deleted successfully")
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
	log.Print("Response Sent")

}

// func respondWithError(w http.ResponseWriter, code int, message string) {
// 	respondWithJSON(w, code, map[string]string{"error": message})
// }

// func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
// 	response, err := json.Marshal(payload)
// 	if err != nil {
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Methods", "POST")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(code)
// 	w.Write(response)
// }
