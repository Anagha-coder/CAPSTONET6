// Function to upload the image to Cloud Storage
package handlers

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"strings"

	"example.com/capstone/utils"
)

// Function to upload the image to Cloud Storage
func ImageAndThumbnailUploadFunc(file multipart.File, item map[string]interface{}) (string, string, error) {
	ctx := context.Background()

	client, err := utils.CreateStorageClient()
	if err != nil {
		log.Print("Failed to create Storage client:", err)
		return "", "", err
	}
	defer client.Close()

	log.Print("Storage client created")

	bucketName := "cloud-storage-bucket-by-anagha"

	productNameValue, ok := item["productName"]
	if !ok || productNameValue == nil {

		return "", "", fmt.Errorf("ProductName is missing or nil")
	}

	weightValue, ok := item["weight"]
	if !ok || weightValue == nil {
		return "", "", fmt.Errorf("weight is missing or nil")
	}

	// Replace spaces with underscores in the product name
	productNameWithoutSpaces := strings.ReplaceAll(productNameValue.(string), " ", "_")

	// Create a unique filename for the image based on the product name and weight
	// Use a suitable format for the weight, e.g., convert to string or format it as needed
	weightStr := fmt.Sprintf("%.2f", weightValue.(float64))
	imageFileName := "images/" + productNameWithoutSpaces + "_" + weightStr + ".jpg"

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
	thumbnailFileName := "thumbnails/" + productNameWithoutSpaces + "_" + weightStr + "_thumbnail.jpg"

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

// func ImageAndThumbnailUploadFunc(file multipart.File, item models.GroceryItem) (string, string, error) {
// 	ctx := context.Background()

// 	client, err := utils.CreateStorageClient()
// 	if err != nil {
// 		log.Print("Failed to create Storage client:", err)
// 		return "", "", err
// 	}
// 	defer client.Close()

// 	log.Print("Storage client created")

// 	bucketName := "cloud-storage-bucket-by-anagha"

// 	// Replace spaces with underscores in the product name
// 	productNameWithoutSpaces := strings.ReplaceAll(item.ProductName, " ", "_")

// 	// Create a unique filename for the image based on the product name and weight
// 	// Use a suitable format for the weight, e.g., convert to string or format it as needed
// 	imageFileName := "images/" + productNameWithoutSpaces + "_" + strconv.FormatFloat(item.Weight, 'f', -1, 64) + ".jpg"

// 	// Create a new GCP Storage object handle
// 	imageObj := client.Bucket(bucketName).Object(imageFileName)

// 	// Create a new writer and upload the image file
// 	imageWC := imageObj.NewWriter(ctx)
// 	if _, err := io.Copy(imageWC, file); err != nil {
// 		return "", "", err
// 	}
// 	if err := imageWC.Close(); err != nil {
// 		return "", "", err
// 	}

// 	// Set the image URL
// 	imageURL := "https://storage.googleapis.com/" + bucketName + "/" + imageFileName

// 	// Reset the file pointer for generating the thumbnail
// 	if _, err := file.Seek(0, io.SeekStart); err != nil {
// 		return "", "", err
// 	}

// 	// Generate thumbnail
// 	thumbnail, err := generateThumbnail(file)
// 	if err != nil {
// 		log.Println("Failed to generate thumbnail:", err)
// 		return "", "", err
// 	}

// 	// Create a unique filename for the thumbnail
// 	thumbnailFileName := "thumbnails/" + productNameWithoutSpaces + "_" + strconv.FormatFloat(item.Weight, 'f', -1, 64) + "_thumbnail.jpg"

// 	// Create a new GCP Storage object handle for the thumbnail
// 	thumbnailObj := client.Bucket(bucketName).Object(thumbnailFileName)

// 	// Create a new writer and upload the thumbnail
// 	thumbnailWC := thumbnailObj.NewWriter(ctx)
// 	if err := jpeg.Encode(thumbnailWC, thumbnail, nil); err != nil {
// 		return "", "", err
// 	}
// 	if err := thumbnailWC.Close(); err != nil {
// 		return "", "", err
// 	}

// 	// Set the thumbnail URL
// 	thumbnailURL := "https://storage.googleapis.com/" + bucketName + "/" + thumbnailFileName

// 	return imageURL, thumbnailURL, nil
// }
