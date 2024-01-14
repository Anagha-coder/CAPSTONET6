package handlers

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"github.com/nfnt/resize"
)

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
	thumbnail := resize.Thumbnail(100, 100, img, resize.Lanczos3)
	if thumbnail == nil {
		log.Println("Generated thumbnail is nil")
		return nil, fmt.Errorf("generated thumbnail is nil")
	}

	log.Println("Image resized, Thumbnail will be returned")
	return thumbnail, nil
}
