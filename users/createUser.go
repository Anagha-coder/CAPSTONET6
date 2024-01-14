package users

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"golang.org/x/crypto/bcrypt"
)

// CreateNewUser creates a new user.
// @Summary Create a new user
// @Description Creates a new user with the provided information
// @ID create-new-user
// @Accept json
// @Produce json
// @Param user body models.User true "User object to be created"
// @Success 201 {object} map[string]string "User Created Successfully. userID"
// @Failure 400 {object} models.ErrorResponse "Invalid request format" or "Missing fields in the request"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error" or "Failed to create user in Firestore"
// @Router /users [post]
func CreateNewUser(w http.ResponseWriter, r *http.Request) {
	// Get the data from the request body

	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	//check if fields are empty

	if newUser.Name == "" || newUser.Email == "" || newUser.Password == "" || newUser.Role == "" {
		respondWithError(w, http.StatusBadRequest, "Missing fields in the request, Check and request again")
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(newUser.Password)
	if err != nil {
		log.Println("Failed to hash password:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// add user to firestore
	userID, err := addUserToFirestore(newUser, hashedPassword)
	if err != nil {
		log.Print("Failed to create grocery item in Firestore:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create grocery item in Firestore")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"User Created Successfully. userID": userID})
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func addUserToFirestore(user models.User, hashedPassword string) (string, error) {

	// Get Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		return "", err
	}
	defer client.Close()

	ctx := context.Background()

	docRef, _, err := client.Collection("users").Add(ctx, map[string]interface{}{
		"Name":     user.Name,
		"Email":    user.Email,
		"Password": hashedPassword,
		"Role":     user.Role,
	})

	if err != nil {
		log.Print("Failed to create user in Firestore:", err)
		return "", err
	}

	return docRef.ID, nil
}
