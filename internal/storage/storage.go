package storage

import (
	"database/sql"
	"embed"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const localDatabaseDSN = "host=localhost user=newuser password=password dbname=gophermart sslmode=disable" // для локальной разработки

type Storage struct {
	db *sql.DB
}

func NewStorage(dsn string) (storage *Storage, err error) {
	if len(dsn) == 0 {
		dsn = localDatabaseDSN
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Register(userID uuid.UUID, login string, password string) (ok bool, err error) {
	// TODO: implementation
	return true, nil
}
