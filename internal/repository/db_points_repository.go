package repository

import (
	"context"

	"github.com/rovany706/loyalty-gopher/internal/database"
	"github.com/rovany706/loyalty-gopher/internal/models"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type DBPointsRepository struct {
	db     *database.Database
	logger *zap.Logger
}

func NewDBPointsRepository(db *database.Database, logger *zap.Logger) *DBPointsRepository {
	return &DBPointsRepository{
		db:     db,
		logger: logger,
	}
}

func (pr *DBPointsRepository) GetUserBalance(ctx context.Context, userID int) (decimal.Decimal, error) {
	var balance decimal.Decimal
	row := pr.db.DBConnection.QueryRowContext(ctx, "SELECT balance FROM point_accounts WHERE user_id=$1", userID)

	err := row.Scan(&balance)

	if err != nil {
		return decimal.Zero, err
	}

	return balance, nil
}

func (pr *DBPointsRepository) GetUserWithdrawalHistory(ctx context.Context, userID int) ([]models.WithdrawHistoryEntry, error) {
	rows, err := pr.db.DBConnection.QueryContext(ctx, `SELECT W.order_num, W.amount, W.processed_at
													   FROM withdrawal_history AS W
													   JOIN point_accounts AS P
													   ON P.id = W.point_account_id
													   WHERE P.user_id=$1
													   ORDER BY W.processed_at DESC`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := make([]models.WithdrawHistoryEntry, 0)

	for rows.Next() {
		var entry models.WithdrawHistoryEntry
		if err := rows.Scan(&entry.OrderNum, &entry.WithdrawSum, &entry.ProcessedAt); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	rerr := rows.Close()
	if rerr != nil {
		return nil, rerr
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (pr *DBPointsRepository) WithdrawPoints(ctx context.Context, userID int, orderNum string, amount decimal.Decimal) error {
	tx, err := pr.db.DBConnection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userPointsAccount struct {
		id      int
		balance decimal.Decimal
	}
	row := tx.QueryRowContext(ctx, "SELECT id, balance FROM point_accounts WHERE user_id=$1", userID)

	err = row.Scan(&userPointsAccount.id, &userPointsAccount.balance)

	if err != nil {
		return err
	}

	if userPointsAccount.balance.LessThan(amount) {
		pr.logger.Info("not enough points", zap.String("balance", userPointsAccount.balance.String()), zap.String("required", amount.String()))
		return ErrNotEnoughPoints
	}

	newBalance := userPointsAccount.balance.Sub(amount)

	_, err = tx.ExecContext(ctx, "UPDATE point_accounts SET balance=$1 WHERE user_id=$2", newBalance, userID)

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO withdrawal_history (order_num, amount, point_account_id) VALUES ($1, $2, $3)", orderNum, amount, userPointsAccount.id)

	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}
