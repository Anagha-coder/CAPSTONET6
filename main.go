package main

import (
	"fmt"
	"net/http"

	_ "example.com/capstone/docs"

	_ "example.com/capstone/models"
	"example.com/capstone/users"

	"example.com/capstone/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	// "github.com/gin-gonic/gin"
)

// @title One Stop Grocery🛒
// @version 1.0
// @description Your ultimate destination for all things fresh, flavorful, and fabulous – where convenience meets quality in our one-stop grocery wonderland!
// @host localhost:8080
// @schemes http
func main() {
	r := mux.NewRouter()

	r.HandleFunc("/createGroceryItem", handlers.CreateGroceryItem).Methods("POST")
	r.HandleFunc("/bulkupload", handlers.BulkUpload).Methods("POST")
	r.HandleFunc("/listGroceryItems", handlers.ListItemsBY).Methods("GET")
	r.HandleFunc("/updateGroceryItemByID/{id:[0-9]+}", handlers.UpdateGroceryItem).Methods("PUT") // impliment patch as well
	r.HandleFunc("/deleteGroceryItemByID/{id:[0-9]+}", handlers.DeleteItemByID).Methods("DELETE")
	r.HandleFunc("/fetchGroceryItemByID/{id:[0-9]+}", handlers.FetchItemByID).Methods("GET")
	r.HandleFunc("/imageUpload", handlers.UploadHandler).Methods("POST")

	// users
	r.HandleFunc("/users", users.CreateNewUser).Methods("POST")
	r.HandleFunc("/userLogin", users.LoginUser).Methods("POST")

	// Swagger UI handler
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // URL to the generated Swagger JSON file
	))

	http.Handle("/", r)

	// Run the server
	fmt.Print("Server is up and running!")
	http.ListenAndServe(":8080", nil)

}

// r := gin.Default()
// r.GET("/ping", func(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// })

// r.POST("/upload", gin.HandlerFunc(handlers.CreateGroceryItem))

// // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
// r.Run()
