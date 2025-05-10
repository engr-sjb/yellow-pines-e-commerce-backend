package product

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/handlerutils"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/middlewares"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/validate"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type servicer interface {
	createProduct(ctx context.Context, newProduct *CreateProductRequest) error
	getAllProducts(ctx context.Context, query *GetAllProductsRequestQuery) ([]*ProductAndInventoryDTO, int, error)
	getProduct(ctx context.Context, productID uuid.UUID) (*ProductAndInventoryDTO, error)
	deleteProduct(ctx context.Context, productID uuid.UUID) error
}

type middleware interface {
	AuthWithContext(h handlerutils.APIHandler, authEntityType string) handlerutils.APIHandler
}

type handler struct {
	service    servicer
	middleware middleware
}

func NewHandler(productService servicer, middleware middleware) *handler {
	return &handler{
		service:    productService,
		middleware: middleware,
	}
}

func (h *handler) RegisterRoutes(router *chi.Mux) {
	router.Get(
		"/products",
		handlerutils.MakeHandler(
			h.getAllProductsHandler,
		),
	)

	router.Get(
		"/products/{productID}",
		handlerutils.MakeHandler(
			h.getProductHandler,
		),
	)

	// protected routes
	router.Post(
		"/products",
		handlerutils.MakeHandler(
			h.middleware.AuthWithContext(
				h.createProductHandler,
				"admin",
			),
		),
	)

}

func (h *handler) createProductHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var payload *CreateProductRequest
	var err error
	defer r.Body.Close()

	if err = handlerutils.ParseJSON(r, &payload); err != nil {
		return servererrors.New(
			http.StatusBadRequest,
			servererrors.ErrInvalidRequestPayload.Error(),
			nil,
		)
	}

	payload.AdminID = middlewares.GetEntityIDFromContextKey(ctx)

	if err = validate.StructFields(payload); err != nil {
		return servererrors.New(
			http.StatusUnprocessableEntity,
			servererrors.ErrValidationFailed.Error(),
			err,
		)
	}

	err = h.service.createProduct(
		ctx,
		payload,
	)
	if err != nil {
		switch {
		case errors.Is(err, servererrors.ErrProductAlreadyExists):
			return servererrors.New(
				http.StatusConflict,
				servererrors.ErrProductAlreadyExists.Error(),
				nil,
			)

		default:
			return err
		}
	}

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusOK,
		"product created",
		nil,
	)
}

func (h *handler) getAllProductsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var err error

	queries := r.URL.Query()

	queryItems, err := getQueryItems(
		queries,
	)
	if err != nil {
		return err
	}

	if err := validate.StructFields(queryItems); err != nil {
		return servererrors.New(
			http.StatusUnprocessableEntity,
			servererrors.ErrURLQueryParams.Error(),
			err,
		)
	}

	products, totalCount, err := h.service.getAllProducts(ctx, queryItems)
	if err != nil {
		return err
	}

	totalPagesCount := totalCount / int(queryItems.PageOpts.Limit)
	itemsLeftCount := (totalCount - int(queryItems.PageOpts.Page*queryItems.PageOpts.Limit))
	pagesLeftCount := (itemsLeftCount + int(queryItems.PageOpts.Limit) - 1) / int(queryItems.PageOpts.Limit)

	if itemsLeftCount < 0 {
		itemsLeftCount = 0
		pagesLeftCount = 0
	}

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusOK,
		"all products retrieved",
		GetAllProductsResponse{
			AllProductsCount:  totalCount,
			RetriedItemsCount: len(products),
			ItemsLeftCount:    itemsLeftCount,
			TotalPagesCount:   totalPagesCount,
			PagesLeftCount:    pagesLeftCount,
			Products:          products,
		},
	)
}

func (h *handler) getProductHandler(w http.ResponseWriter, r *http.Request) error {
	productIDStr := chi.URLParam(r, "productID")

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return err
	}
	product, err := h.service.getProduct(r.Context(), productID)
	if err != nil {
		return err
	}

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusOK,
		"product found",
		product,
	)
}


func getQueryItems(queriesParams url.Values) (*GetAllProductsRequestQuery, error) {
	query := new(GetAllProductsRequestQuery)

	query.FilterOpts.Category = queriesParams.Get("category")

	results := strings.Split(queriesParams.Get("sort"), ":")

	query.SortOpts.SortBy = "created_at"
	query.SortOpts.SortOpt = "desc"

	if len(results) == 1 && results[0] != "" {
		query.SortOpts.SortBy = results[0]
	}

	if len(results) == 2 {
		query.SortOpts.SortBy = results[0]
		query.SortOpts.SortOpt = results[1]
	}

	query.FilterOpts.Search = queriesParams.Get("search")

	query.PageOpts.Page = stringToUint64(
		1,
		queriesParams.Get("page"),
	)

	query.PageOpts.Limit = stringToUint64(
		20,
		queriesParams.Get("limit"),
	)

	query.FilterOpts.PriceMin = stringToFloat64(
		0.00,
		queriesParams.Get("priceMin"),
	)

	query.FilterOpts.PriceMax = stringToFloat64(
		0.00,
		queriesParams.Get("priceMax"),
	)

	return query, nil
}

func stringToUint64(defaultValue uint64, field string) uint64 {
	num, err := strconv.ParseUint(field, 10, 0)
	if err != nil {
		return defaultValue
	}

	return num
}

func stringToFloat64(defaultValue float64, field string) float64 {
	num, err := strconv.ParseFloat(field, 64)
	if err != nil {
		return defaultValue
	}

	return num
}
