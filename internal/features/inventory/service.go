package inventory

import (
	"context"

	"github.com/google/uuid"
)

type storer interface {
	createOne(ctx context.Context, pdID uuid.UUID, stkQty uint) error
}

type service struct {
	store storer
}

func NewService(inventoryStore storer) *service {
	return &service{
		store: inventoryStore,
	}
}

func (s *service) CreateInventory(ctx context.Context, pdID uuid.UUID, stkQty uint) error {
	return s.store.createOne(ctx, pdID, stkQty)
}
