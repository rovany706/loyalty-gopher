package database

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const MigrationSQL = `
DROP TABLE IF EXISTS withdrawal_history;
DROP TABLE IF EXISTS point_accounts;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS e_accrual_status;

CREATE TYPE e_accrual_status AS ENUM (
    'REGISTERED',
    'INVALID',
    'PROCESSED',
    'PROCESSING'
);

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username TEXT UNIQUE NOT NULL,
    pw_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_num TEXT UNIQUE NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    accrual_status e_accrual_status NOT NULL,
    accrual NUMERIC(12,2) NOT NULL,
    user_id INT REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS point_accounts (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    balance NUMERIC(12,2) NOT NULL,
    user_id INT REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS withdrawal_history (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_num TEXT NOT NULL,
    amount NUMERIC(12,2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    point_account_id INT REFERENCES point_accounts(id)
);
`

type Database struct {
	DBConnection *sql.DB
}

func InitConnection(ctx context.Context, databaseURI string) (*Database, error) {
	dbConnection, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return nil, err
	}

	db := &Database{
		DBConnection: dbConnection,
	}

	return db, nil
}

func (db *Database) EnsureCreated(ctx context.Context) error {
	_, err := db.DBConnection.ExecContext(ctx, MigrationSQL)

	return err
}

// not working in github workflow
// func (db *Database) RunMigrations(ctx context.Context) error {
// 	driver, err := pgx.WithInstance(db.DBConnection, &pgx.Config{})

// 	if err != nil {
// 		return err
// 	}

// 	m, err := migrate.NewWithDatabaseInstance("file://../../internal/database/migrations", "", driver)

// 	if err != nil {
// 		return err
// 	}

// 	err = m.Up()

// 	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
// 		return err
// 	}

// 	return nil
// }
