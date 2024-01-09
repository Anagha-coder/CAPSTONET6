package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"google.golang.org/api/iterator"
)

func BulkUpload(w http.ResponseWriter, r *http.Request) {
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

	// Get the file from the form data
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Println("Failed to get file:", err)
		respondWithError(w, http.StatusBadRequest, "Failed to get file")
		return
	}
	defer file.Close()

	// Determine the file type (CSV or JSON) based on content type or file extension
	fileType, err := determineFileType(file)
	if err != nil {
		log.Println("Failed to determine file type:", err)
		respondWithError(w, http.StatusBadRequest, "Failed to determine file type")
		return
	}

	// Read and process the file based on its type
	var groceryItems []models.GroceryItem
	switch fileType {
	case "csv":
		groceryItems, err = readCSVFile(file)
	case "json":
		groceryItems, err = readJSONFile(file)
	default:
		respondWithError(w, http.StatusBadRequest, "Unsupported file type")
		return
	}

	if err != nil {
		log.Println("Failed to process file:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process file")
		return
	}

	// Iterate through the grocery items and perform the necessary actions
	for _, groceryItem := range groceryItems {
		// Create a Firestore client
		client, err := utils.CreateFirestoreClient()
		if err != nil {
			log.Print("Failed to create Firestore client:", err)
			// Handle the error if needed
			continue
		}
		defer client.Close()

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
				// Handle the error if needed
				continue
			}
			var item models.GroceryItem
			if err := doc.DataTo(&item); err != nil {
				log.Print("Failed to parse grocery item data from Firestore:", err)
				// Handle the error if needed
				continue
			}
			existingGroceryItems = append(existingGroceryItems, item)
		}

		// Generate a unique ID for the new grocery item
		newItemID := generateUniqueGroceryItemID(existingGroceryItems)

		// Set the new grocery item ID
		groceryItem.ID = newItemID

		// Set the Firestore document ID to the product name
		docRef := client.Collection("groceryItems").Doc(groceryItem.ProductName)

		// Add the new grocery item to Firestore with the specified document ID
		_, err = docRef.Set(context.Background(), groceryItem)
		if err != nil {
			log.Print("Failed to create grocery item in Firestore:", err)
			// Handle the error if needed
			continue
		}

		log.Print("Grocery item created successfully in Firestore")
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "Bulk grocery items created successfully"})
	log.Print("Response Sent: CreateBulkGroceryItems")
}

func determineFileType(file multipart.File) (string, error) {
	// Read the first 512 bytes to determine the file type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Check for CSV file signature
	if isCSV(buffer) {
		return "csv", nil
	}

	// Check for JSON file signature
	if isJSON(buffer) {
		return "json", nil
	}

	// If the file type is not recognized, return an error
	return "", errors.New("unsupported file type")
}

func isCSV(buffer []byte) bool {
	// Check if the file starts with the CSV signature
	csvSignature := []byte{0xEF, 0xBB, 0xBF} // CSV files may have a UTF-8 BOM
	return bytes.HasPrefix(buffer, csvSignature) || strings.HasPrefix(string(buffer), "sep=") || strings.Contains(string(buffer), ",")
}

func isJSON(buffer []byte) bool {
	// Check if the file starts with the JSON signature
	return bytes.HasPrefix(buffer, []byte{'{', '['})
}

func readCSVFile(file multipart.File) ([]models.GroceryItem, error) {
	// Create a CSV reader
	csvReader := csv.NewReader(file)

	// Read all records from the CSV file
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Process CSV records and convert them to GroceryItem objects
	var groceryItems []models.GroceryItem
	for _, record := range records {
		groceryItem := models.GroceryItem{
			ProductName: record[0],
			// Category:record[1],
			// Price:record[2],
			// Weight:record[3],
			// WeightUnit: record[4], // Adjust the index based on your CSV structure
			// Populate other fields accordingly
		}
		groceryItems = append(groceryItems, groceryItem)
	}

	return groceryItems, nil
}

func readJSONFile(file multipart.File) ([]models.GroceryItem, error) {
	// Decode the JSON file
	decoder := json.NewDecoder(file)

	// Decode JSON array into a slice of GroceryItem
	var groceryItems []models.GroceryItem
	if err := decoder.Decode(&groceryItems); err != nil {
		return nil, err
	}

	return groceryItems, nil
}
