package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rovany706/loyalty-gopher/internal/database"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type DBOrderRepository struct {
	db     *database.Database
	logger *zap.Logger
}

func NewDBOrderRepository(db *database.Database, logger *zap.Logger) *DBOrderRepository {
	return &DBOrderRepository{
		db:     db,
		logger: logger,
	}
}

func (r *DBOrderRepository) GetOrder(ctx context.Context, orderNum string) (*models.Order, error) {
	row := r.db.DBConnection.QueryRowContext(ctx, "SELECT order_num, user_id, uploaded_at, accrual_status, accrual FROM orders WHERE order_num=$1", orderNum)
	var order models.Order
	err := row.Scan(&order.OrderNum, &order.UserID, &order.UploadedAt, &order.AccrualStatus, &order.Accrual)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

func (r *DBOrderRepository) AddOrder(ctx context.Context, userID int, orderNum string) error {
	_, err := r.db.DBConnection.ExecContext(ctx, "INSERT INTO orders (order_num, user_id, accrual_status, accrual) VALUES ($1, $2, $3, 0)", orderNum, userID, models.AccrualStatusRegistered)

	r.logger.Info("added order", zap.String("num", orderNum))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = ErrOrderConflict
		}
	}

	return err
}

func (r *DBOrderRepository) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	rows, err := r.db.DBConnection.QueryContext(ctx, "SELECT order_num, user_id, uploaded_at, accrual_status, accrual FROM orders WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]models.Order, 0)

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.OrderNum, &order.UserID, &order.UploadedAt, &order.AccrualStatus, &order.Accrual); err != nil {
			return nil, err
		}

		if order.Accrual.IsZero() {
			order.Accrual = nil
		}

		orders = append(orders, order)
	}

	rerr := rows.Close()
	if rerr != nil {
		return nil, rerr
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *DBOrderRepository) UpdateOrderStatus(ctx context.Context, orderNum string, newAccrualStatus models.AccrualStatus, accrualAmount *decimal.Decimal) error {
	tx, err := r.db.DBConnection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if accrualAmount == nil {
		r.logger.Info("accrual is nil")
		accrualAmount = &decimal.Zero
	}
	_, err = tx.ExecContext(ctx, "UPDATE orders SET accrual_status=$1, accrual=$2 WHERE order_num=$3", newAccrualStatus, *accrualAmount, orderNum)

	if err != nil {
		r.logger.Info("err1", zap.Error(err))

		return err
	}

	err = tx.Commit()

	if err != nil {
		r.logger.Info("err2", zap.Error(err))
		return err
	}

	return nil
}
