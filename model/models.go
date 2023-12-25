package model

import (
	"time"
)

type GroceryItem struct {
	ProductName         string    `json:"productName" validate:"required"`
	Category            string    `json:"category" validate:"required"`
	Price               float64   `json:"price" validate:"required"`
	Weight              float64   `json:"weight" validate:"required"`
	Vegetarian          bool      `json:"vegetarian"`
	Image               string    `json:"image" validate:"required"` // stored on bucket, req
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
	Password  string
	// Role ? Admin OR manager

}

// productName
//price
// weight
// manufacturer
// category
// img
// vegeterian
// itemPackagequantity
// PackageInformation
//Country of origin
//Mfg date ? & Exp date

// schema for user
// auth based on userid and password
// ID
// passeord
// role?
