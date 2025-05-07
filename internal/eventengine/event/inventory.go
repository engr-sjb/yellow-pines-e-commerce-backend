package event

import "github.com/google/uuid"

const (
	InventoryCreationFailedEventName EventName = "inventory.creation.failed"
)

type InventoryCreationFailedEvent struct {
	ProductID uuid.UUID
}

func (e *InventoryCreationFailedEvent) GetEventName() EventName {
	return InventoryCreationFailedEventName
}
