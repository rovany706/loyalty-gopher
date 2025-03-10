package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database struct {
	DBConnection *sql.DB
}

func InitConnection(ctx context.Context, databaseUri string) (*Database, error) {
	dbConnection, err := sql.Open("pgx", databaseUri)
	if err != nil {
		return nil, err
	}

	db := &Database{
		DBConnection: dbConnection,
	}

	return db, nil
}

func (db *Database) RunMigrations(ctx context.Context) error {
	driver, err := pgx.WithInstance(db.DBConnection, &pgx.Config{})

	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://../../internal/database/migrations", "", driver)

	if err != nil {
		return err
	}

	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
