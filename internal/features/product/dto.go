package product

import (
	"github.com/google/uuid"
)

// Requests

type CreateProductRequest struct {
	AdminID     uuid.UUID `json:"adminID"`
	Name        string    `json:"name" validate:"required,min=10,max=30,noAllRepeatingChars"`
	Description string    `json:"description" validate:"required,min=15,max=350,noAllRepeatingChars"`
	ImageURL    string    `json:"imageURL" validate:"required,url"`
	Price       float64   `json:"price" validate:"required,gt=0"`
	Category    string    `json:"category" validate:"required"`
	Quantity    uint      `json:"quantity" validate:"required"`
}

type UpdateProductRequest struct {
	AdminID     uuid.UUID `json:"adminID" validate:"required,uuid"`
	ProductID   string    `json:"productID" validate:"required,uuid"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	ImageURL    *string   `json:"imageURL"`
	Price       *float64  `json:"price"`
	Category    *string   `json:"category"`
	IsActive    *bool     `json:"isActive"`
	Quantity    *uint     `json:"quantity" validate:"required"`
}

type FilterOpts struct {
	Category string  `json:"category"`
	PriceMin float64 `json:"priceMin" validate:"min=0"`
	PriceMax float64 `json:"priceMax" validate:"min=0"`
	Search   string  `json:"search"`
}

type SortOpts struct {
	SortBy  string `json:"sortBy" validate:"oneof=name price category created_at"`
	SortOpt string `json:"sortOpt" validate:"oneof=desc asc"`
}

type PageOpts struct {
	Page  uint64 `json:"page" validate:"min=0"`
	Limit uint64 `json:"limit" validate:"min=0"`
}

type GetAllProductsRequestQuery struct {
	FilterOpts FilterOpts `json:"filterOpts"`
	SortOpts   SortOpts   `json:"sortOpts"`
	PageOpts   PageOpts   `json:"pageOpts"`
}

// Responses

type ProductAndInventoryDTO struct {
	Product
	StockQuantity uint `json:"stockQuantity"`
}

type GetAllProductsResponse struct {
	AllProductsCount  int                       `json:"allProductsCount"`
	RetriedItemsCount int                       `json:"retriedItemsCount"`
	TotalPagesCount   int                       `json:"totalPagesCount"`
	PagesLeftCount    int                       `json:"pagesLeftCount"`
	ItemsLeftCount    int                       `json:"itemsLeftCount"`
	Products          []*ProductAndInventoryDTO `json:"products"`
}
