package handlers

type GroceryItem struct {
	ProductName         string  `json:"productName" `
	Category            string  `json:"category" `
	Price               float64 `json:"price" `
	Weight              float64 `json:"weight" `
	WeightUnit          string  `json:"weightUnit" `
	Vegetarian          bool    `json:"vegetarian"`
	Manufacturer        string  `json:"manufacturer" `
	Brand               string  `json:"brand" `
	ItemPackageQuantity int     `json:"itemPackageQuantity" `
	PackageInformation  string  `json:"packageInformation" `
	MfgDate             struct {
		Month int `json:"month"`
		Year  int `json:"year"`
	} `json:"mfgDate" `
	ExpDate struct {
		Month int `json:"month"`
		Year  int `json:"year"`
	} `json:"expDate" `
	CountryOfOrigin string `json:"countryOfOrigin" `
}
