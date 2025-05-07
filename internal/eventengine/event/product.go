package event

import "github.com/google/uuid"

var (
	ProductCreatedEventName         EventName = "product.created"
	ProductUpdatedEventName         EventName = "product.updated"
	ProductUpdatedQuantityEventName EventName = "product.updated.quantity"
	ProductDeletedEventName         EventName = "product.deleted"
)

type ProductPayload struct {
	ProductID     uuid.UUID
	StockQuantity uint
}

type ProductCreatedEvent struct {
	ProductPayload
}

func (e ProductCreatedEvent) GetEventName() EventName {
	return ProductCreatedEventName
}

type ProductUpdatedEvent struct {
	Name EventName
	ProductPayload
}

func (e ProductUpdatedEvent) EventName() EventName {
	return ProductUpdatedEventName
}

type ProductQuantityUpdatedEvent struct {
	Name EventName
	ProductPayload
}

func (e ProductQuantityUpdatedEvent) EventName() EventName {
	return ProductUpdatedQuantityEventName
}

type ProductDeletedEvent struct {
	Name      EventName
	ProductID uuid.UUID
}

func (e ProductDeletedEvent) EventName() EventName {
	return ProductDeletedEventName
}
