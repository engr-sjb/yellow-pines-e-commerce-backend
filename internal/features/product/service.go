package product

import (
	"context"
	"strings"

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

type inventoryServicer interface {
	CreateInventory(ctx context.Context, pdID uuid.UUID, stkQty uint) error
}

type service struct {
	store            storer
	inventoryService inventoryServicer
}

func NewService(productStore storer, inventoryService inventoryServicer) *service {
	return &service{
		store:            productStore,
		inventoryService: inventoryService,
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

	if err := s.inventoryService.CreateInventory(
		ctx,
		pdID,
		newProduct.Quantity,
	); err != nil {
		fErr := err
		if err := s.store.deleteOne(ctx, pdID); err != nil {
			return fmt.Errorf(
				"error deleting product after inventory creation failed. inventory: %w, product: %w", // todo: revisit and rework err
				fErr,
				err,
			)
		}
	}

	return nil
}

func (s *service) getAllProducts(ctx context.Context, queryItems *GetAllProductsRequestQuery) ([]*ProductAndInventoryDTO, int, error) {
	return s.store.findAll(ctx, queryItems)
}

func (s *service) getProduct(ctx context.Context, productID uuid.UUID) (*ProductAndInventoryDTO, error) {
	return s.store.findByID(ctx, productID)
}

