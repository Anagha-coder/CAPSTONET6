package models

import (
	"time"
)

type GroceryItem struct {
	ID                  int       `json:"id"`
	ProductName         string    `json:"productName" validate:"required"`
	Category            string    `json:"category" validate:"required"`
	Price               float64   `json:"price" validate:"required"`
	Weight              float64   `json:"weight" validate:"required"`
	WeightUnit          string    `json:"weightUnit" validate:"required"` // e.g., "gm", "kg", "ml", "l"
	Vegetarian          bool      `json:"vegetarian"`
	Image               string    `json:"imageURL" validate:"required"` // stored on bucket, req  - datatype - []string - to store image names in it as ref
	ImageHash           string    `json:"imageHash" firestore:"imageHash"`
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

type User struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Password  string // not string hashed password
	Role      string // Role ? Admin OR manager - one role for now - admin

}

type AuditRecord struct {
	Action    string    `json:"action"`
	ItemID    string    `json:"itemID"`
	Timestamp time.Time `json:"timestamp"`
	// PerformedBy string    `json:"performedBy"`
}
