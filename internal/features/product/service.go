package product

import (
	"context"
	"strings"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/google/uuid"
)

type Storer interface {
	createOne(ctx context.Context, product *CreateProductRequest) error
	findAll(ctx context.Context, queryItems *GetAllProductsRequestQuery) ([]*ProductAndInventoryDTO, int, error)
	findByID(ctx context.Context, productID uuid.UUID) (*ProductAndInventoryDTO, error)
	findByName(ctx context.Context, name string) (*Product, error)
}

type service struct {
	store Storer
}

func NewService(store Storer) *service {
	return &service{
		store: store,
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

	return s.store.createOne(
		ctx,
		newProduct,
	)
}

func (s *service) getAllProducts(ctx context.Context, queryItems *GetAllProductsRequestQuery) ([]*ProductAndInventoryDTO, int, error) {
	return s.store.findAll(ctx, queryItems)
}

func (s *service) getProduct(ctx context.Context, productID uuid.UUID) (*ProductAndInventoryDTO, error) {
	return s.store.findByID(ctx, productID)
}

