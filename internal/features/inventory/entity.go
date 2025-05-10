package inventory

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ProductID        uuid.UUID `json:"productID"`
	StockQuantity    uint      `json:"stockQuantity"`
	RestockThreshold uint      `json:"restockThreshold"`
	UpdatedAt        time.Time `json:"updatedAt"`
	ReservedQuantity uint      `json:"reservedQuantity"`
}
