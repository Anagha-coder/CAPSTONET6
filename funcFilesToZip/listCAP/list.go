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
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
)

func init() {
	functions.HTTP("ListItemsBY", ListItemsBY)
}

// ListItemsBY lists grocery items based on query parameters.
// @Summary List grocery items based on query parameters
// @Description Retrieves a list of grocery items from the Firestore database based on the provided query parameters.
// @ID list-items-by
// @Produce json
// @Param productName query string false "Filter by product name"
// @Param price query number false "Filter by price"
// @Param price_min query number false "Filter by minimum price"
// @Param price_max query number false "Filter by maximum price"
// @Param Category query string false "Filter by category"
// @Param pageSize query integer false "Number of items per page" format(int32)
// @Param pageNumber query integer false "Page number" format(int32)
// @Success 201 {Object} string "List Of Grocery Items"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not Found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /listGroceryItems [get]
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

		if k == "pageSize" || k == "pageNumber" {
			continue
		}
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

	// Calculate pagination parameters
	pageSizeStr := r.URL.Query().Get("pageSize")
	pageNumberStr := r.URL.Query().Get("pageNumber")

	var pageSize, pageNumber int

	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			log.Printf("Error parsing pageSize: %v", err)
		}
	} else {
		// default page size
		pageSize = 10
	}

	if pageNumberStr != "" {
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			log.Printf("Error parsing pageNumber: %v", err)
		}
	} else {
		// default page no
		pageNumber = 1
	}

	// Calculate the start index for pagination
	startIndex := (pageNumber - 1) * pageSize

	log.Printf("pageSize: %d, pageNumber: %d, startIndex: %d", pageSize, pageNumber, startIndex)

	// Add pagination to the query
	query = query.Offset(0).Limit(pageSize)

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
