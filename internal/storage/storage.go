package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/iubondar/gophermart/internal/storage/queries"
	"github.com/iubondar/gophermart/internal/validator"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Storage struct {
	db *sql.DB
}

func NewStorage(dsn string) (storage *Storage, err error) {
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

func (s *Storage) Register(ctx context.Context, userID uuid.UUID, login string, passwordHash string) (ok bool, err error) {
	_, err = s.db.ExecContext(ctx, queries.InsertUser, userID, login, passwordHash)
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

func (s *Storage) CheckLogin(ctx context.Context, login string, password string) (userID uuid.UUID, err error) {
	row := s.db.QueryRowContext(ctx, queries.GetUserID, login)

	var hashedPassword string
	err = row.Scan(&userID, &hashedPassword)

	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return uuid.Nil, nil
	}

	return userID, nil
}

func (s *Storage) RegisterOrder(ctx context.Context, userID uuid.UUID, orderNumber string) (result constants.OrderRegistrationResult, err error) {
	// Валидируем номер заказа алгоритмом Луна
	isValid := validator.ValidateLuhn(orderNumber)
	if !isValid {
		return constants.WrongOrderNumberFormat, nil
	}

	// Ищем заказ по номеру
	var orderUserID uuid.UUID
	row := s.db.QueryRowContext(ctx, queries.GetUserIDByOrderNum, orderNumber)
	err = row.Scan(&orderUserID)
	if errors.Is(err, sql.ErrNoRows) {
		// Не нашли  - сохраняем заказ в БД с начальным статусом
		_, err = s.db.ExecContext(ctx, queries.RegisterOrder, userID, orderNumber, constants.OrderStatusNew)
		if err != nil {
			zap.L().Sugar().Debugln("Error insert new order:", err.Error())
			return 0, err
		}
		return constants.AcceptedToProcessing, nil
	} else if err != nil {
		// Другая ошибка - возвращаем ошибку
		zap.L().Sugar().Debugln("Error query order by num:", err.Error())
		return 0, err
	} else {
		// Нашли - сверяем userID и возвращаем нужный результат
		if userID == orderUserID {
			return constants.AlreadyRegistered, nil
		} else {
			return constants.RegisteredByAnotherUser, nil
		}
	}
}

func (s *Storage) Orders(ctx context.Context, userID uuid.UUID) (orders []models.Order, err error) {
	rows, err := s.db.QueryContext(ctx, queries.Orders, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.Order{}, nil
		} else {
			return nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing rows: %s", err.Error())
	}

	return orders, nil
}

func (s *Storage) OrdersToUpdate(ctx context.Context, limit int) (orders []models.Order, err error) {
	rows, err := s.db.QueryContext(ctx, queries.OrdersToUpdate, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return orders, nil
		} else {
			return nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing rows: %s", err.Error())
	}

	return orders, nil
}

func (s *Storage) UpdateOrders(ctx context.Context, orders []models.OrderStatus) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// если Commit будет раньше, то откат проигнорируется
	defer tx.Rollback()

	updateStmt, err := tx.PrepareContext(ctx, queries.UpdateOrder)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	addStmt, err := tx.PrepareContext(ctx, queries.AddBalance)
	if err != nil {
		return err
	}
	defer addStmt.Close()

	for _, order := range orders {
		// обновляем статус заказа
		_, err = updateStmt.ExecContext(ctx, order.Status, order.Accrual, order.Number)
		if err != nil {
			return err
		}

		// добавляем баланс пользователю, если заказ обработан
		if order.Status == constants.OrderStatusProcessed {
			_, err = addStmt.ExecContext(ctx, order.Accrual, order.UserID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *Storage) Withdrawals(ctx context.Context, userID uuid.UUID) (withdrawals []models.Withdrawal, err error) {
	rows, err := s.db.QueryContext(ctx, queries.Withdrawals, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.Withdrawal{}, nil
		} else {
			return nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		var withdrawal models.Withdrawal
		err = rows.Scan(&withdrawal.Number, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing rows: %s", err.Error())
	}

	return withdrawals, nil
}

func (s *Storage) Withdraw(ctx context.Context, userID uuid.UUID, orderNumber string, sum float32) (result constants.WithdrawResult, err error) {
	// Валидируем номер заказа алгоритмом Луна
	isValid := validator.ValidateLuhn(orderNumber)
	if !isValid {
		return constants.WrongOrderFormat, nil
	}

	// Проверяем баланс пользователя
	row := s.db.QueryRowContext(ctx, queries.GetBalance, userID)

	var balance float32
	err = row.Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("cannot find record for userID: %s", userID.String())
		} else {
			return 0, err
		}
	}
	if balance < sum {
		return constants.InsufficientFunds, nil
	}

	// Уменьшаем баланс на полученную сумму и делаем запись о списании в одной транзакции
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	withdrawFromBalanceStmt, err := tx.PrepareContext(ctx, queries.WithdrawFromBalance)
	if err != nil {
		return 0, err
	}
	defer withdrawFromBalanceStmt.Close()

	addWithdrawStmt, err := tx.PrepareContext(ctx, queries.AddWithdraw)
	if err != nil {
		return 0, err
	}
	defer addWithdrawStmt.Close()

	_, err = withdrawFromBalanceStmt.ExecContext(ctx, sum, userID)
	if err != nil {
		return 0, err
	}

	_, err = addWithdrawStmt.ExecContext(ctx, userID, orderNumber, sum)
	if err != nil {
		return 0, err
	}

	return constants.Success, tx.Commit()
}

func (s *Storage) Account(ctx context.Context, userID uuid.UUID) (account models.Account, err error) {
	// Баланс пользователя
	row := s.db.QueryRowContext(ctx, queries.GetBalance, userID)

	var balance float32
	err = row.Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Account{}, fmt.Errorf("cannot find record for userID: %s", userID.String())
		} else {
			return models.Account{}, err
		}
	}

	// Сумма вывода
	row = s.db.QueryRowContext(ctx, queries.WithdrawalSum, userID)

	var withdrawalSum float32
	err = row.Scan(&withdrawalSum)
	if err != nil {
		return models.Account{}, err
	}

	return models.Account{
		Balance:       balance,
		WithdrawalSum: withdrawalSum,
	}, nil
}
