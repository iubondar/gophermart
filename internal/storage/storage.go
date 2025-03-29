package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/storage/queries"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
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

func (s *Storage) Register(ctx context.Context, userID uuid.UUID, login string, password_hash string) (ok bool, err error) {
	_, err = s.db.ExecContext(ctx, queries.InsertUser, userID, login, password_hash)
	if err != nil {
		// Если пользователь с логином уже существует - возвращаем не ок
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return false, nil
		}

		// Другая ошибка
		zap.L().Sugar().Debugln("Error insert new user:", err.Error())
		return false, err
	}

	return true, nil
}
