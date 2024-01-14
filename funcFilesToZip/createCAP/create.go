package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/dgrijalva/jwt-go"
	"github.com/nfnt/resize"
	"google.golang.org/api/iterator"
)

type GroceryItem struct {
	ID                  int       `json:"id"`
	ProductName         string    `json:"productName" validate:"required"`
	Category            string    `json:"category" validate:"required"`
	Price               float64   `json:"price" validate:"required"`
	Weight              float64   `json:"weight" validate:"required"`
	WeightUnit          string    `json:"weightUnit" validate:"required"` // e.g., "gm", "kg", "ml", "l"
	Vegetarian          bool      `json:"vegetarian"`
	Image               string    `json:"imageURL" validate:"required"`     // stored on bucket, req  - datatype - []string - to store image names in it as ref
	Thumbnail           string    `json:"thumbnailURL" validate:"required"` // stored on bucket, req  - datatype - []string - to store image names in it as ref
	Manufacturer        string    `json:"manufacturer" validate:"required"`
	Brand               string    `json:"brand" validate:"required"`
	ItemPackageQuantity int       `json:"itemPackageQuantity" validate:"required"`
	PackageInformation  string    `json:"packageInformation" validate:"required"`
	MfgDate             MonthYear `json:"mfgDate" validate:"required"`
	ExpDate             MonthYear `json:"expDate" validate:"required"`
	CountryOfOrigin     string    `json:"countryOfOrigin" validate:"required"`
}

type MonthYear struct {
	Month time.Month
	Year  int
}

func init() {
	functions.HTTP("CreateGroceryItem", CreateGroceryItem)
}

const (
	tokenSecret = "jdcklqnfioqwfnhipndklasxnesrtsrx"
	// Replace with your actual secret key
)

// CreateGroceryItem creates a new grocery item.

// @Summary Create a new grocery item
// @Description Creates a new grocery item and uploads its image to your database. Image is optional, you can add it later by using update method as well. Do provide 'Bearer' before adding authorization token
// @ID create-grocery-item
// @Accept json
// @Produce json
// @Param Authorization header string true "token"
// @Param json-data formData string  true "JSON data for the grocery item" format(json) x-example({"name": "Example Item", "quantity": 10})
// @Param image formData file false "Optional: Image file for the grocery item"
// @Success 201 {string} string "Grocery item created successfully"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Not Found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /createGroceryItem [post]
// @Security BearerToken
func CreateGroceryItem(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	utils.InitLogger()

	// Extract the token from the request header
	tokenString := ExtractToken(r)
	if tokenString == "" {
		respondWithError(w, http.StatusUnauthorized, "Token not provided")
		return
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Check if the token is valid and not expired
	if !token.Valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Check if the token is valid and not expired
	// if !token.Valid {
	// 	// Attempt to refresh the token
	// 	refreshedToken, refreshErr := RefreshToken(tokenString)
	// 	if refreshErr != nil {
	// 		respondWithError(w, http.StatusUnauthorized, "Invalid token")
	// 		return
	// 	}

	// 	// Update the response with the new token
	// 	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Token refreshed successfully", "newToken": refreshedToken})
	// 	return
	// }

	// Create a Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create Firestore client")
		return
	}
	defer client.Close()

	log.Print("Firestore client created")

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
	var groceryItem models.GroceryItem
	if err := json.Unmarshal([]byte(jsonData), &groceryItem); err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// image file from the form data
	file, _, err := r.FormFile("image")
	// log.Printf("Original image format: %s", formatimg)
	if err == http.ErrMissingFile {
		// if no image provided, proceed without image
		log.Println("No image file")
	} else if err != nil {
		log.Println("Failed to get image file:", err)
		respondWithError(w, http.StatusBadRequest, "Failed to get image file")
		return
	} else {
		defer file.Close()

		// Reset the file pointer to the beginning
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			log.Println("Failed to reset file pointer:", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to reset file pointer")
			return
		}

		// // Read the first 512 bytes to detect the content type
		// buffer := make([]byte, 512)
		// _, err := file.Read(buffer)
		// if err != nil {
		// 	log.Println("Failed to read file content:", err)
		// 	respondWithError(w, http.StatusBadRequest, "Failed to read file content")
		// 	return
		// }

		// // Check if the content type is image based on the detected content
		// contentType := http.DetectContentType(buffer)
		// if !strings.HasPrefix(contentType, "image/") {
		// 	log.Println("Invalid file type. Only images are allowed.")
		// 	respondWithError(w, http.StatusBadRequest, "Invalid file type. Only images are allowed.")
		// 	return
		// }

		// Calculate SHA-256 hash of the image data
		imageHash, err := CalculateImageHash(file)
		if err != nil {
			log.Println("Failed to calculate image hash:", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to calculate image hash")
			return
		}

		log.Println("Image hash calculated successfully:", imageHash)

		// Check if an item with the same hash already exists
		if IsDuplicateImage(imageHash) {
			log.Println("Duplicate image detected")
			respondWithError(w, http.StatusBadRequest, "Duplicate image detected")
			return
		}

		// Use a channel to signal when the image upload is complete
		imageUploadDone := make(chan bool)

		// a goroutine to handle image upload asynchronously
		go func() {
			// Upload the image to the Cloud Storage bucket
			imageURL, thumbnailURL, err := uploadImageAndThumbailToCloudStorage(file, groceryItem)
			if err != nil {
				log.Println("failed to upload thumbnail to cloud storage:", err)
				// Handle error if needed
				imageUploadDone <- false
				return
			}

			// Set the image URL in the grocery item
			groceryItem.Image = imageURL
			groceryItem.Thumbnail = thumbnailURL
			groceryItem.ImageHash = imageHash

			// Signal that the image upload is complete
			imageUploadDone <- true
		}()

		// Wait for the image upload to complete (or for a timeout)
		select {
		case <-imageUploadDone:
			log.Println("Image uploaded!")
		case <-time.After(30 * time.Second):
			log.Println("Image upload timed out")

		}

	}

	// Read existing grocery items from Firestore - collection : "groceryItems"
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

	// a unique ID for the new grocery item
	newItemID := generateUniqueGroceryItemID(existingGroceryItems)

	// Set the new grocery item ID
	groceryItem.ID = newItemID

	// Set the Firestore document ID to the product name
	// docRef := client.Collection("groceryItems").Doc(groceryItem.ProductName)

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

// Function to upload the image to Cloud Storage
func uploadImageAndThumbailToCloudStorage(file multipart.File, item models.GroceryItem) (string, string, error) {
	ctx := context.Background()

	client, err := utils.CreateStorageClient()
	if err != nil {
		log.Print("Failed to create Storage client:", err)
		return "", "", err
	}
	defer client.Close()

	log.Print("Storage client created")

	bucketName := "cloud-storage-bucket-by-anagha"

	// Replace spaces with underscores in the product name
	productNameWithoutSpaces := strings.ReplaceAll(item.ProductName, " ", "_")

	// Create a unique filename for the image based on the product name and weight
	// Use a suitable format for the weight, e.g., convert to string or format it as needed
	imageFileName := "images/" + productNameWithoutSpaces + "_" + strconv.FormatFloat(item.Weight, 'f', -1, 64) + ".jpg"

	// Create a new GCP Storage object handle
	imageObj := client.Bucket(bucketName).Object(imageFileName)

	// Create a new writer and upload the image file
	imageWC := imageObj.NewWriter(ctx)
	if _, err := io.Copy(imageWC, file); err != nil {
		return "", "", err
	}
	if err := imageWC.Close(); err != nil {
		return "", "", err
	}

	// Set the image URL
	imageURL := "https://storage.googleapis.com/" + bucketName + "/" + imageFileName

	// Reset the file pointer for generating the thumbnail
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}

	// Generate thumbnail
	thumbnail, err := generateThumbnail(file)
	if err != nil {
		log.Println("Failed to generate thumbnail:", err)
		return "", "", err
	}

	// Create a unique filename for the thumbnail
	thumbnailFileName := "thumbnails/" + productNameWithoutSpaces + "_" + strconv.FormatFloat(item.Weight, 'f', -1, 64) + "_thumbnail.jpg"

	// Create a new GCP Storage object handle for the thumbnail
	thumbnailObj := client.Bucket(bucketName).Object(thumbnailFileName)

	// Create a new writer and upload the thumbnail
	thumbnailWC := thumbnailObj.NewWriter(ctx)
	if err := jpeg.Encode(thumbnailWC, thumbnail, nil); err != nil {
		return "", "", err
	}
	if err := thumbnailWC.Close(); err != nil {
		return "", "", err
	}

	// Set the thumbnail URL
	thumbnailURL := "https://storage.googleapis.com/" + bucketName + "/" + thumbnailFileName

	return imageURL, thumbnailURL, nil
}

func generateThumbnail(file io.Reader) (image.Image, error) {
	// Decode the original image
	log.Println("Before decoding the image")
	img, format, err := image.Decode(file)
	log.Printf("Original image format: %s", format)
	if err != nil {
		log.Printf("Error decoding the original image: %v", err)
		log.Printf("Original image format: %s", format)
		return nil, err
	}
	log.Println("After decoding the image")

	// Resize the image to create a thumbnail
	thumbnail := resize.Thumbnail(500, 500, img, resize.Lanczos3)
	if thumbnail == nil {
		log.Println("Generated thumbnail is nil")
		return nil, fmt.Errorf("generated thumbnail is nil")
	}

	log.Println("Image resized, Thumbnail will be returned")
	return thumbnail, nil
}
