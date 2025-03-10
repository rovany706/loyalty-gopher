package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rovany706/loyalty-gopher/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type DBUserRepository struct {
	db *database.Database
}

func NewDBUserRepository(db *database.Database) *DBUserRepository {
	return &DBUserRepository{
		db: db,
	}
}

func (r *DBUserRepository) Register(ctx context.Context, login string, password string) (userID int, err error) {
	hashedPassword, err := hashPassword(password)

	if err != nil {
		return UnauthorizedUserID, err
	}

	row := r.db.DBConnection.QueryRowContext(ctx, "INSERT INTO users (username, pw_hash) VALUES ($1, $2) RETURNING id", login, hashedPassword)

	err = row.Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return UnauthorizedUserID, ErrUserConfict
		}

		return UnauthorizedUserID, err
	}

	return userID, nil
}

func (r *DBUserRepository) Login(ctx context.Context, login string, password string) (userID int, err error) {
	row := r.db.DBConnection.QueryRowContext(ctx, "SELECT id, pw_hash FROM users WHERE username=$1", login)
	var userInfo struct {
		id             int
		hashedPassword string
	}
	err = row.Scan(&userInfo.id, &userInfo.hashedPassword)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UnauthorizedUserID, nil
		}

		return UnauthorizedUserID, err
	}

	isPasswordCorrect := checkPassword(password, userInfo.hashedPassword)

	if !isPasswordCorrect {
		return UnauthorizedUserID, nil
	}

	return userInfo.id, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
