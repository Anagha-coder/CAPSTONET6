package handlers

import (
	"bytes"
	"context"

	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"

	"example.com/capstone/utils"
)

// BulkUpload uploads a file containing grocery items in CSV or JSON format.
// @Summary Upload a file with grocery items
// @Description Uploads a file containing grocery items in CSV or JSON format to the server, which then processes and stores the items
// @ID bulk-upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File containing grocery items (CSV or JSON)"
// @Success 201 {object} map[string]string "File Uploaded to Cloud Storage"
// @Failure 400 {object} ErrorResponse "Invalid request format" or "Failed to parse multipart form" or "Failed to determine file type" or "Failed to get file"
// @Failure 500 {object} ErrorResponse "Failed to create Storage client" or "Failed to upload file to cloud storage"
// @Router /bulkupload [post]
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
	file, fileHeader, err := r.FormFile("file")
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

	// Log the detected file type
	log.Printf("Detected file type: %v", fileType)

	storageClient, err := utils.CreateStorageClient()
	if err != nil {
		log.Print("Failed to create Storage client:", err)
		return
	}
	defer storageClient.Close()

	// Get the original filename
	originalFilename := fileHeader.Filename

	// Upload the file to the cloud storage bucket with the original filename
	cloudStoragePath := "dataFiles/" + originalFilename
	if err := uploadToCloudStorage(storageClient, cloudStoragePath, file); err != nil {
		log.Println("Failed to upload file to cloud storage:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to upload file to cloud storage")
		return
	}

	log.Printf("File uploaded successfully to %s", cloudStoragePath)
	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "File Uploaded to Cloud Storage"})
	log.Print("Response Sent: CreateBulkGroceryItems")
}

// Function to upload the file to cloud storage
func uploadToCloudStorage(client *storage.Client, cloudStoragePath string, file io.Reader) error {
	ctx := context.Background()

	// Open a bucket handle
	bucket := client.Bucket("cloudbucketanaghaaaa")
	obj := bucket.Object(cloudStoragePath)

	// Create a writer to upload the file
	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		wc.Close()
		return err
	}

	// Close the writer to finalize the upload
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
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
	var js map[string]interface{}
	err := json.Unmarshal(buffer, &js)
	if err != nil {
		log.Printf("JSON Unmarshal error: %v", err)
	}

	return err == nil
}
