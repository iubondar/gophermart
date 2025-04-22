package storage

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/iubondar/gophermart/internal/testhelpers"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type StorageTestSuite struct {
	suite.Suite
	storage *Storage
	cleanup func()
}

func (s *StorageTestSuite) SetupSuite() {
	ctx := context.Background()
	container, err := testhelpers.CreatePostgresContainer(ctx)
	require.NoError(s.T(), err)

	db, err := sql.Open("pgx", container.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	goose.SetDialect("postgres")
	err = goose.Up(db, "./migrations")
	if err != nil {
		log.Fatal(err)
	}

	storage, err := NewStorage(container.ConnectionString)
	require.NoError(s.T(), err)

	s.storage = storage
	s.cleanup = func() {
		container.Terminate(ctx)
	}
}

func (s *StorageTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *StorageTestSuite) TearDownTest() {
	ctx := context.Background()
	_, err := s.storage.db.ExecContext(ctx, "TRUNCATE TABLE users CASCADE")
	require.NoError(s.T(), err)
	_, err = s.storage.db.ExecContext(ctx, "TRUNCATE TABLE orders CASCADE")
	require.NoError(s.T(), err)
	_, err = s.storage.db.ExecContext(ctx, "TRUNCATE TABLE withdrawals CASCADE")
	require.NoError(s.T(), err)
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (s *StorageTestSuite) TestRegister() {
	ctx := context.Background()
	userID := uuid.New()
	login := "testuser"
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(s.T(), err)

	s.Run("successful registration", func() {
		ok, err := s.storage.Register(ctx, userID, login, string(passwordHash))
		assert.NoError(s.T(), err)
		assert.True(s.T(), ok)
	})

	s.Run("duplicate registration", func() {
		ok, err := s.storage.Register(ctx, userID, login, string(passwordHash))
		assert.NoError(s.T(), err)
		assert.False(s.T(), ok)
	})
}

func (s *StorageTestSuite) TestCheckLogin() {
	ctx := context.Background()
	userID := uuid.New()
	login := "testuser"
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(s.T(), err)

	// Register user first
	ok, err := s.storage.Register(ctx, userID, login, string(passwordHash))
	require.NoError(s.T(), err)
	require.True(s.T(), ok)

	s.Run("successful login", func() {
		resultID, err := s.storage.CheckLogin(ctx, login, password)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), userID, resultID)
	})

	s.Run("wrong password", func() {
		resultID, err := s.storage.CheckLogin(ctx, login, "wrongpassword")
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), uuid.Nil, resultID)
	})

	s.Run("non-existent user", func() {
		resultID, err := s.storage.CheckLogin(ctx, "nonexistent", password)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), uuid.Nil, resultID)
	})
}

func (s *StorageTestSuite) TestRegisterOrder() {
	ctx := context.Background()
	userID := uuid.New()
	orderNumber := "12345678903" // Valid Luhn number

	// Register user first
	ok, err := s.storage.Register(ctx, userID, "testuser", "password")
	require.NoError(s.T(), err)
	require.True(s.T(), ok)

	s.Run("successful order registration", func() {
		result, err := s.storage.RegisterOrder(ctx, userID, orderNumber)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), constants.AcceptedToProcessing, result)
	})

	s.Run("duplicate order registration", func() {
		result, err := s.storage.RegisterOrder(ctx, userID, orderNumber)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), constants.AlreadyRegistered, result)
	})

	s.Run("invalid order number", func() {
		result, err := s.storage.RegisterOrder(ctx, userID, "12345678901")
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), constants.WrongOrderNumberFormat, result)
	})
}

func (s *StorageTestSuite) TestOrders() {
	ctx := context.Background()
	userID := uuid.New()
	orderNumber := "12345678903"

	// Register user and order
	_, err := s.storage.Register(ctx, userID, "testuser", "password")
	require.NoError(s.T(), err)
	_, err = s.storage.RegisterOrder(ctx, userID, orderNumber)
	require.NoError(s.T(), err)

	s.Run("get user orders", func() {
		orders, err := s.storage.Orders(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Len(s.T(), orders, 1)
		assert.Equal(s.T(), orderNumber, orders[0].Number)
		assert.Equal(s.T(), constants.OrderStatusNew, orders[0].Status)
	})

	s.Run("get non-existent user orders", func() {
		orders, err := s.storage.Orders(ctx, uuid.New())
		assert.NoError(s.T(), err)
		assert.Empty(s.T(), orders)
	})
}

func (s *StorageTestSuite) TestWithdrawals() {
	ctx := context.Background()
	userID := uuid.New()
	orderNumber := "12345678903"
	sum := float32(100.0)

	// Register user and add balance
	ok, err := s.storage.Register(ctx, userID, "testuser", "password")
	require.NoError(s.T(), err)
	require.True(s.T(), ok)

	// Add balance to user
	_, err = s.storage.db.ExecContext(ctx, "UPDATE users SET balance = $1 WHERE user_id = $2", sum*2, userID)
	require.NoError(s.T(), err)

	s.Run("get withdrawals for user with no withdrawals", func() {
		withdrawals, err := s.storage.Withdrawals(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Empty(s.T(), withdrawals)
	})

	s.Run("get withdrawals after withdrawal", func() {
		// Make a withdrawal
		result, err := s.storage.Withdraw(ctx, userID, orderNumber, sum)
		require.NoError(s.T(), err)
		require.Equal(s.T(), constants.Success, result)

		withdrawals, err := s.storage.Withdrawals(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Len(s.T(), withdrawals, 1)
		assert.Equal(s.T(), orderNumber, withdrawals[0].Number)
		assert.Equal(s.T(), sum, withdrawals[0].Sum)
		assert.NotZero(s.T(), withdrawals[0].ProcessedAt)
	})
}

func (s *StorageTestSuite) TestAccount() {
	ctx := context.Background()
	userID := uuid.New()
	balance := float32(100.0)

	// Register user and set balance
	_, err := s.storage.Register(ctx, userID, "testuser", "password")
	require.NoError(s.T(), err)

	_, err = s.storage.db.ExecContext(ctx, "UPDATE users SET balance = $1 WHERE user_id = $2",
		balance, userID)
	require.NoError(s.T(), err)

	s.Run("get account info", func() {
		account, err := s.storage.Account(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), balance, account.Balance)
		assert.Equal(s.T(), float32(0), account.WithdrawalSum) // Initially 0
	})

	s.Run("get non-existent account", func() {
		_, err := s.storage.Account(ctx, uuid.New())
		assert.Error(s.T(), err)
	})

	s.Run("get account with withdrawals", func() {
		// Make a withdrawal
		orderNumber := "12345678903"
		sum := float32(50.0)
		result, err := s.storage.Withdraw(ctx, userID, orderNumber, sum)
		require.NoError(s.T(), err)
		require.Equal(s.T(), constants.Success, result)

		account, err := s.storage.Account(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), balance-sum, account.Balance)
		assert.Equal(s.T(), sum, account.WithdrawalSum)
	})
}

func (s *StorageTestSuite) TestOrdersToUpdate() {
	ctx := context.Background()
	userID := uuid.New()
	orderNumber := "12345678903"

	// Register user and order
	_, err := s.storage.Register(ctx, userID, "testuser", "password")
	require.NoError(s.T(), err)
	_, err = s.storage.RegisterOrder(ctx, userID, orderNumber)
	require.NoError(s.T(), err)

	s.Run("get orders to update", func() {
		orders, err := s.storage.OrdersToUpdate(ctx, 10)
		assert.NoError(s.T(), err)
		assert.Len(s.T(), orders, 1)
		assert.Equal(s.T(), orderNumber, orders[0].Number)
		assert.Equal(s.T(), constants.OrderStatusNew, orders[0].Status)
	})

	s.Run("get orders to update with limit", func() {
		// Add more orders
		orderNumber2 := "12345678904"
		_, err = s.storage.RegisterOrder(ctx, userID, orderNumber2)
		require.NoError(s.T(), err)

		orders, err := s.storage.OrdersToUpdate(ctx, 1)
		assert.NoError(s.T(), err)
		assert.Len(s.T(), orders, 1)
	})
}

func (s *StorageTestSuite) TestUpdateOrders() {
	ctx := context.Background()
	userID := uuid.New()
	orderNumber := "12345678903"
	accrual := float32(100.0)

	// Register user and order
	_, err := s.storage.Register(ctx, userID, "testuser", "password")
	require.NoError(s.T(), err)
	_, err = s.storage.RegisterOrder(ctx, userID, orderNumber)
	require.NoError(s.T(), err)

	s.Run("update order status", func() {
		orders := []models.OrderStatus{
			{
				UserID:  userID,
				Number:  orderNumber,
				Status:  constants.OrderStatusProcessed,
				Accrual: accrual,
			},
		}

		err := s.storage.UpdateOrders(ctx, orders)
		assert.NoError(s.T(), err)

		// Verify order status and balance
		account, err := s.storage.Account(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), accrual, account.Balance)

		userOrders, err := s.storage.Orders(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Len(s.T(), userOrders, 1)
		assert.Equal(s.T(), constants.OrderStatusProcessed, userOrders[0].Status)
		assert.Equal(s.T(), accrual, userOrders[0].Accrual)
	})

	s.Run("update multiple orders", func() {
		// Clear balance first
		_, err = s.storage.db.ExecContext(ctx, "UPDATE users SET balance = 0 WHERE user_id = $1", userID)
		require.NoError(s.T(), err)

		// Add another order
		orderNumber2 := "12345678904"
		accrual2 := float32(50.0)
		_, err = s.storage.RegisterOrder(ctx, userID, orderNumber2)
		require.NoError(s.T(), err)

		orders := []models.OrderStatus{
			{
				UserID:  userID,
				Number:  orderNumber,
				Status:  constants.OrderStatusProcessed,
				Accrual: accrual,
			},
			{
				UserID:  userID,
				Number:  orderNumber2,
				Status:  constants.OrderStatusProcessed,
				Accrual: accrual2,
			},
		}

		err := s.storage.UpdateOrders(ctx, orders)
		assert.NoError(s.T(), err)

		// Verify total balance
		account, err := s.storage.Account(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), accrual+accrual2, account.Balance)
	})
}

func (s *StorageTestSuite) TestWithdraw() {
	ctx := context.Background()
	userID := uuid.New()
	orderNumber := "12345678903"
	sum := float32(100.0)

	// Register user and add balance
	ok, err := s.storage.Register(ctx, userID, "testuser", "password")
	require.NoError(s.T(), err)
	require.True(s.T(), ok)

	// Add balance to user
	_, err = s.storage.db.ExecContext(ctx, "UPDATE users SET balance = $1 WHERE user_id = $2", sum*2, userID)
	require.NoError(s.T(), err)

	s.Run("successful withdrawal", func() {
		result, err := s.storage.Withdraw(ctx, userID, orderNumber, sum)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), constants.Success, result)

		// Verify balance was updated
		account, err := s.storage.Account(ctx, userID)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), sum, account.Balance) // Initial balance was sum*2, withdrew sum
	})

	s.Run("insufficient funds", func() {
		result, err := s.storage.Withdraw(ctx, userID, orderNumber, sum*3)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), constants.InsufficientFunds, result)
	})

	s.Run("invalid order number", func() {
		result, err := s.storage.Withdraw(ctx, userID, "12345678901", sum)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), constants.WrongOrderFormat, result)
	})
}
