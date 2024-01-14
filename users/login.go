package users

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/capstone/models"
	"example.com/capstone/utils"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
)

const (
	tokenSecret       = "jdcklqnfioqwfnhipndklasxnesrtsrx" // Change this to a strong secret key
	tokenExpiration   = 24 * time.Hour                     // Token expiration time
	tokenRefreshDelta = 30 * time.Minute                   // Time before expiration to refresh token
)

// LoginUser logs in a user and generates a JWT token

// @Summary Log in a user
// @Description Authenticates a user with the provided email and password, returning a JWT token upon success
// @ID login-user
// @Accept json
// @Produce json
// @Param loginCredentials body models.LoginUser true "Login credentials (email and password)"
// @Success 200 {object} map[string]string "Token"
// @Failure 400 {object} models.ErrorResponse "Invalid request format"
// @Failure 401 {object} models.ErrorResponse "Invalid email or password"
// @Failure 500 {object} models.ErrorResponse "Failed to generate token"
// @Router /userLogin [post]
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginUser models.LoginUser
	if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate login credentials
	user, err := authenticateUser(loginUser.Email, loginUser.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		log.Println("Failed to generate token:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Display token in response
	respondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

func authenticateUser(email, password string) (*models.User, error) {
	// Get Firestore client
	client, err := utils.CreateFirestoreClient()
	if err != nil {
		log.Print("Failed to create Firestore client:", err)
		return nil, err
	}
	defer client.Close()

	ctx := context.Background()

	// Query user by email
	query := client.Collection("users").Where("Email", "==", email)
	iter := query.Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Print("Failed to iterate user documents:", err)
			return nil, err
		}

		var user models.User
		doc.DataTo(&user)

		// Check password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err == nil {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found or invalid password")
}

func generateToken(user *models.User) (string, error) {
	// Create a new token with the user's ID as the subject
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(tokenExpiration).Unix(),
	})

	// Sign the token with a secret key
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func refreshToken(w http.ResponseWriter, r *http.Request) {
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

	// Check if the token is close to expiration
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Invalid token claims")
		return
	}

	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	timeUntilExpiration := expirationTime.Sub(time.Now())

	if timeUntilExpiration > tokenRefreshDelta {
		respondWithError(w, http.StatusBadRequest, "Token not close to expiration")
		return
	}

	// Refresh the token
	newToken, err := generateTokenFromClaims(claims)
	if err != nil {
		log.Println("Failed to refresh token:", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	// Display the new token in the response
	respondWithJSON(w, http.StatusOK, map[string]string{"token": newToken})
}

func generateTokenFromClaims(claims jwt.MapClaims) (string, error) {
	// Update the expiration time
	claims["exp"] = time.Now().Add(tokenExpiration).Unix()

	// Create a new token with the updated claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ExtractToken(r *http.Request) string {
	// Get the token from the Authorization header
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return ""
	}

	// Extract the token part
	tokenParts := strings.Split(bearerToken, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		return ""
	}

	return tokenParts[1]
}
