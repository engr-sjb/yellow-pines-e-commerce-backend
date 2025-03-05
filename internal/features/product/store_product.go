package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) createOne(ctx context.Context, product *CreateProductRequest) error {
	tx, err := s.db.BeginTx(
		ctx,
		nil,
	)
	if err != nil {
		return err
	}

	productQuery := `INSERT INTO products(admin_id, name, description, image_url, price, category) VALUES($1, $2, $3, $4, $5, $6) RETURNING product_id`

	var productID uuid.UUID

	err = tx.QueryRowContext(
		ctx,
		productQuery,
		product.AdminID,
		product.Name,
		product.Description,
		product.ImageURL,
		product.Price,
		product.Category,
	).Scan(&productID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(
			"failed to insert new product in product store: %w",
			err,
		)
	}

	inventoryQuery := `INSERT INTO inventory(product_id, stock_quantity) VALUES($1, $2)`
	_, err = tx.ExecContext(
		ctx,
		inventoryQuery,
		productID,
		product.Quantity,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(
			"failed to insert new product into inventory in product store: %w",
			err,
		)
	}

	return tx.Commit()
}

func (s *Store) findAll(ctx context.Context, queryItems *GetAllProductsRequestQuery) (products []*Product, count int, err error) {
	query, countQuery, queryParams := generateQueryAndParams(
		queryItems,
	)

	err = s.db.QueryRowContext(
		ctx,
		countQuery,
		queryParams[:len(queryParams)-2]..., // exclude limit and offset
	).Scan(
		&count,
	)
	if err != nil {
		return nil, 0, fmt.Errorf(
			"failed to get all products count from product store: %w",
			err,
		)
	}

	rows, err := s.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, 0, fmt.Errorf(
			"failed to get all products from product store: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ProductID,
			&product.AdminID,
			&product.Name,
			&product.Description,
			&product.ImageURL,
			&product.Price,
			&product.Category,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf(
				"failed to scan product from product store: %w",
				err,
			)
		}
		products = append(products, &product)
	}

	return products, count, nil
}

func (s *Store) findByID(ctx context.Context, productID uuid.UUID) (*Product, error) {
	query := `SELECT * FROM products WHERE product_id = $1`
	row := s.db.QueryRowContext(ctx, query, productID)
	var product Product
	err := row.Scan(
		&product.ProductID,
		&product.AdminID,
		&product.Name,
		&product.Description,
		&product.ImageURL,
		&product.Price,
		&product.Category,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt)
	if err != nil {
		return &product, fmt.Errorf(
			"failed to scan product from product store: %w",
			err,
		)
	}

	return &product, nil
}

func (s *Store) findByName(ctx context.Context, name string) (*Product, error) {
	query := `SELECT * FROM products WHERE name = $1`
	rows, err := s.db.QueryContext(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	product := new(Product)
	for rows.Next() {
		err = scanRowsIntoProduct(rows, product)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return product, nil
			}

			return product, fmt.Errorf(
				"/product store/: failed to scan into product: %w",
				err,
			)
		}
	}

	return product, nil
}


func scanRowsIntoProduct(rows *sql.Rows, product *Product) error {
	return rows.Scan(
		&product.ProductID,
		&product.AdminID,
		&product.Name,
		&product.Description,
		&product.ImageURL,
		&product.Price,
		&product.Category,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
}

func generateQueryAndParams(queryItems *GetAllProductsRequestQuery) (string, string, []any) {
	// Base SQL query
	defaultQuery := `SELECT * FROM products`
	defaultCountQuery := `SELECT COUNT(*) FROM products`

	whereClauses := []string{}
	queryParams := []any{}
	sortClause := ""
	// selectFields := "*" // Default to all fields

	if queryItems.FilterOpts.Search != "" {
		whereClauses = append(
			whereClauses,
			fmt.Sprintf(
				"(name ILIKE $%d OR description ILIKE $%d)",
				len(queryParams)+1, len(queryParams)+2,
			),
		)

		queryParams = append(
			queryParams,
			fmt.Sprintf(
				"%s%%",
				queryItems.FilterOpts.Search,
			),
			fmt.Sprintf(
				"%s%%",
				queryItems.FilterOpts.Search,
			))
	}

	if queryItems.FilterOpts.Category != "" {
		whereClauses = append(
			whereClauses,
			fmt.Sprintf(
				"category = $%d",
				len(queryParams)+1,
			),
		)

		queryParams = append(queryParams, queryItems.FilterOpts.Category)
	}

	if queryItems.FilterOpts.PriceMin > 0.00 {
		whereClauses = append(
			whereClauses,
			fmt.Sprintf(
				"price >= $%d",
				len(queryParams)+1,
			),
		)
		queryParams = append(queryParams, queryItems.FilterOpts.PriceMin)
	}

	if queryItems.FilterOpts.PriceMax > 0.00 {
		whereClauses = append(
			whereClauses,
			fmt.Sprintf("price <= $%d", len(queryParams)+1),
		)

		queryParams = append(queryParams, queryItems.FilterOpts.PriceMax)
	}

	if queryItems.SortOpts.SortBy != "" {
		// **Important Security Note:** Be very careful with dynamic ORDER BY clauses in raw SQL.
		//  Ensure `queryItems.SortOpts.SortBy` is validated against a whitelist of allowed columns
		//  to prevent SQL injection vulnerabilities. For simplicity in this example, we're assuming it's validated.
		sortClause = fmt.Sprintf(
			"ORDER BY %s %s",
			queryItems.SortOpts.SortBy,
			strings.ToUpper(queryItems.SortOpts.SortOpt),
		)
	}

	// --- Construct queries ---
	if len(whereClauses) > 0 {
		whereStr := strings.Join(whereClauses, " AND ")

		defaultQuery += fmt.Sprintf(
			" WHERE %s",
			whereStr,
		)

		defaultCountQuery += fmt.Sprintf(
			" WHERE %s",
			whereStr,
		)
	}

	if sortClause != "" {
		defaultQuery += fmt.Sprintf(" %s", sortClause)
	}

	// --- Pagination LIMIT and OFFSET ---
	defaultQuery += fmt.Sprintf(
		" LIMIT $%d OFFSET $%d",
		len(queryParams)+1,
		len(queryParams)+2,
	)
	queryParams = append(
		queryParams,
		queryItems.PageOpts.Limit,
		(queryItems.PageOpts.Page-1)*queryItems.PageOpts.Limit,
	)

	return defaultQuery, defaultCountQuery, queryParams
}
