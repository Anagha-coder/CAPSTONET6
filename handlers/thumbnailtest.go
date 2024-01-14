package handlers

import (
	"fmt"
	"image"

	// "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/nfnt/resize"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form data with a max of 10 MB limit for the entire request
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println("Failed to parse multipart form:", err)
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the image file from the form data
	file, _, err := r.FormFile("image")
	if err != nil {
		log.Println("Failed to get image file:", err)
		http.Error(w, "Failed to get image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a thumbnail from the image
	thumbnail, err := createThumbnail(file)
	if err != nil {
		log.Println("Failed to create thumbnail:", err)
		http.Error(w, "Failed to create thumbnail", http.StatusInternalServerError)
		return
	}

	// Save the thumbnail to a file (for demonstration purposes)
	saveThumbnailToFile(thumbnail, "thumbnail.png")

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Thumbnail created successfully")
}

func createThumbnail(file io.Reader) (image.Image, error) {
	// Decode the original image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding the original image: %v", err)
	}

	// Resize the image to create a thumbnail
	thumbnail := resize.Thumbnail(1000, 1000, img, resize.Lanczos3)
	if thumbnail == nil {
		return nil, fmt.Errorf("generated thumbnail is nil")
	}

	return thumbnail, nil
}

func saveThumbnailToFile(thumbnail image.Image, filename string) error {
	// Create a new file for the thumbnail
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating thumbnail file: %v", err)
	}
	defer file.Close()

	// Encode the thumbnail and write it to the file
	err = png.Encode(file, thumbnail)
	if err != nil {
		return fmt.Errorf("error encoding thumbnail: %v", err)
	}

	return nil
}
