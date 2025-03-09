package product

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ProductID   uuid.UUID `json:"productID"`
	AdminID     uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"imageURL"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Inventory struct {
	ProductID        uuid.UUID `json:"productID"`
	StockQuantity    uint      `json:"stockQuantity"`
	RestockThreshold uint      `json:"restockThreshold"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
