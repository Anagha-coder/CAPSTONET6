package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"google.golang.org/api/iterator"
)

func CreateBulkGroceryItems(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	utils.InitLogger()

	// Parse the form data with a max of 10 MB limit for the entire request
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println("Failed to parse multipart form:", err)
		respondWithError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}
	log.Println("File will be sent on the request by User")

	// Get the file from the form data
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Println("Failed to get file:", err)
		respondWithError(w, http.StatusBadRequest, "Failed to get file")
		return
	}
	defer file.Close()
	log.Println("File received")

	// Read and process the file
	var groceryItems []models.GroceryItem
	switch strings.ToLower(strings.TrimSpace(r.FormValue("filetype"))) {
	case "csv":
		groceryItems, err = readGroceryItemsFromCSV(file)
		log.Println("Read from CSV file")
		if err != nil {
			log.Println("Failed to read grocery items from CSV:", err)
			respondWithError(w, http.StatusBadRequest, "Failed to read grocery items from CSV")
			return

		}

	case "json":
		groceryItems, err = readGroceryItemsFromJSON(file)
		if err != nil {
			log.Println("Failed to read grocery items from JSON:", err)
			respondWithError(w, http.StatusBadRequest, "Failed to read grocery items from JSON")
			return
		}
		// default:
		// 	respondWithError(w, http.StatusBadRequest, "Invalid file type. Supported types: CSV, JSON")

	}

	// Create a Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create Firestore client")
		return
	}
	defer client.Close()
	log.Println("Firestore client created")

	// Read existing grocery items from Firestore (assuming collection named "groceryItems")
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
	log.Printf("Read %d existing grocery items from Firestore", len(existingGroceryItems))
	log.Println("Before the loop")

	// Iterate over the bulk grocery items and add them to Firestore
	for _, item := range groceryItems {
		// Generate a unique ID for the new grocery item
		newItemID := generateUniqueGroceryItemID(existingGroceryItems)
		item.ID = newItemID
		log.Printf("Generated new item ID: %d", newItemID)

		// Set the Firestore document ID to the product name
		docRef := client.Collection("groceryItems").Doc(item.ProductName)
		log.Printf("Adding new grocery item to Firestore. Product Name: %s, Document ID: %s", item.ProductName, docRef.ID)

		// Add the new grocery item to Firestore with the specified document ID
		_, err := docRef.Set(context.Background(), item)
		if err != nil {
			log.Printf("Failed to create grocery item '%s' in Firestore: %v", item.ProductName, err)
			// Handle error if needed
		} else {
			log.Printf("Created grocery item '%s' in Firestore", item.ProductName)
		}
		log.Println("after the loop")

	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "Bulk grocery items created successfully"})
	log.Print("Response Sent: CreateBulkGroceryItems")
}

func readGroceryItemsFromCSV(file io.Reader) ([]models.GroceryItem, error) {
	var groceryItems []models.GroceryItem

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Read all records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Assuming the CSV has a header row
	header := records[0]

	// Iterate over the remaining rows (data rows)
	for _, row := range records[1:] {
		item := models.GroceryItem{}

		// Map each field in the row to the corresponding struct field
		for i, value := range row {
			setStructField(&item, header[i], value)
		}

		// Append the item to the slice
		groceryItems = append(groceryItems, item)
	}

	return groceryItems, nil
}

func readGroceryItemsFromJSON(file io.Reader) ([]models.GroceryItem, error) {
	var groceryItems []models.GroceryItem

	decoder := json.NewDecoder(file)
	for {
		var item models.GroceryItem
		if err := decoder.Decode(&item); err == io.EOF {
			log.Println("End of JSON file reached")
			break
		} else if err != nil {
			log.Printf("Error decoding JSON: %v", err)
			return nil, err
		}

		groceryItems = append(groceryItems, item)
		log.Printf("Read grocery item from JSON: %+v", item)
	}

	return groceryItems, nil
}

func setStructField(item *models.GroceryItem, fieldName, value string) {
	// Use reflection to set the struct field based on the field name
	structValue := reflect.ValueOf(item).Elem()
	fieldValue := structValue.FieldByName(fieldName)

	if fieldValue.IsValid() {
		switch fieldValue.Kind() {
		case reflect.Int:
			intValue, err := strconv.Atoi(value)
			if err == nil {
				fieldValue.SetInt(int64(intValue))
			}
		case reflect.Float64:
			floatValue, err := strconv.ParseFloat(value, 64)
			if err == nil {
				fieldValue.SetFloat(floatValue)
			}
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err == nil {
				fieldValue.SetBool(boolValue)
			}
		case reflect.String:
			fieldValue.SetString(value)
			// Add more cases for other field types as needed
		}
	}
}
