package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	tokenExpiration   = 24 * time.Hour // Token expiration time
	tokenRefreshDelta = 30 * time.Minute
)

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

func RefreshToken(w http.ResponseWriter, r *http.Request) {
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
