package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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

func (db *Database) Close() error {
	return db.DBConnection.Close()
}

func (db *Database) RunMigrations(ctx context.Context) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	driver, err := pgx.WithInstance(db.DBConnection, &pgx.Config{})

	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "", driver)
	if err != nil {
		return err
	}

	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
