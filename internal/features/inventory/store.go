package inventory

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *store {
	return &store{
		db: db,
	}
}

func (s *store) createOne(ctx context.Context, pdID uuid.UUID, stkQty uint) error {
	inventoryQuery := `INSERT INTO inventory(product_id, stock_quantity) VALUES($1, $2)`

	_, err := s.db.ExecContext(
		ctx,
		inventoryQuery,
		pdID,
		stkQty,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to insert new product into inventory in product store: %w",
			err,
		)
	}
	return nil
}
