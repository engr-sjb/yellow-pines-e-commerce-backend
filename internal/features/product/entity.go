package product

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ProductID   uuid.UUID `json:"product_id"`
	AdminID     uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"-"`
}

type Inventory struct {
	ProductID        uuid.UUID `json:"product_id"`
	StockQuantity    uint      `json:"stock_quantity"`
	RestockThreshold uint      `json:"restock_threshold"`
	UpdatedAt        time.Time `json:"updated_at"`
}
