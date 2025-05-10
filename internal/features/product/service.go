package product

import (
	"context"
	"fmt"
	"strings"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine/event"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/google/uuid"
)

type storer interface {
	createOne(ctx context.Context, product *CreateProductRequest) (uuid.UUID, error)
	findAll(ctx context.Context, queryItems *GetAllProductsRequestQuery) ([]*ProductAndInventoryDTO, int, error)
	findByID(ctx context.Context, pdID uuid.UUID) (*ProductAndInventoryDTO, error)
	findByName(ctx context.Context, name string) (*Product, error)
	deleteOne(ctx context.Context, pdID uuid.UUID) error
}

type service struct {
	store storer
	// inventoryService inventoryServicer // todo: remove this and replace with event engine. or keep it.
	eventEngine eventengine.Publisher
}

func NewService(productStore storer, eventEngine eventengine.Publisher) *service {
	return &service{
		store:       productStore,
		eventEngine: eventEngine,
	}
}

func (s *service) createProduct(ctx context.Context, newProduct *CreateProductRequest) error {
	newProduct.Name = strings.TrimSpace(newProduct.Name)
	newProduct.Description = strings.TrimSpace(newProduct.Description)
	newProduct.ImageURL = strings.TrimSpace(newProduct.ImageURL)

	product, err := s.store.findByName(ctx, newProduct.Name)
	if err != nil {
		return err
	}

	if product.ProductID != uuid.Nil {
		return servererrors.ErrProductAlreadyExists
	}

	pdID, err := s.store.createOne(
		ctx,
		newProduct,
	)
	if err != nil {
		return err
	}

	// if err := s.inventoryService.CreateInventory(
	// 	ctx,
	// 	pdID,
	// 	newProduct.Quantity,
	// ); err != nil {
	// 	fErr := err
	// 	if err := s.store.deleteOne(ctx, pdID); err != nil {
	// 		return fmt.Errorf(
	// 			"error deleting product after inventory creation failed. inventory: %w, product: %w", // todo: revisit or keep it.
	// 			fErr,
	// 			err,
	// 		)
	// 	}
	// }

	newEvent := &event.ProductCreatedEvent{
		ProductPayload: event.ProductPayload{
			ProductID:     pdID,
			StockQuantity: newProduct.Quantity,
		},
	}

	err = s.eventEngine.Publish(
		&event.Event{
			Name:    newEvent.GetEventName(),
			Payload: newEvent,
		},
	)
	if err != nil {
		firstErr := err

		if err := s.store.deleteOne(ctx, pdID); err != nil {
			return fmt.Errorf(
				"error deleting newly created product after publishing event failed. eventEnginePublish: %w, product: %w",
				firstErr,
				err,
			)
		}

		return fmt.Errorf(
			"error publishing event after product creation. eventEnginePublish: %w",
			firstErr,
		)
	}

	return nil
}

func (s *service) getAllProducts(ctx context.Context, queryItems *GetAllProductsRequestQuery) ([]*ProductAndInventoryDTO, int, error) {
	return s.store.findAll(ctx, queryItems)
}

func (s *service) getProduct(ctx context.Context, productID uuid.UUID) (*ProductAndInventoryDTO, error) {
	return s.store.findByID(ctx, productID)
}

func (s *service) deleteProduct(ctx context.Context, productID uuid.UUID) error {
	err := s.store.deleteOne(
		ctx,
		productID,
	)

	if err != nil {
		return fmt.Errorf(
			"err happened: %w",
			err,
		)
	}

	return nil
}
