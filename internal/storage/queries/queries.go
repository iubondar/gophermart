package queries

const (
	InsertUser string = "INSERT INTO users (user_id, user_name, password_hash) VALUES ($1, $2, $3);"

	GetUserID string = "SELECT user_id, password_hash from users WHERE user_name = $1;"

	GetUserIDByOrderNum string = "SELECT user_id from orders WHERE order_number = $1;"

	RegisterOrder string = "INSERT INTO orders (user_id, order_number, order_status) VALUES ($1, $2, $3);"

	Orders string = "SELECT order_number, order_status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC;"

	OrdersToUpdate string = `SELECT user_id, order_number, order_status, accrual, uploaded_at  FROM orders 
	WHERE order_status IN ('NEW', 'PROCESSING')
	ORDER BY uploaded_at ASC LIMIT $1`

	UpdateOrder string = "UPDATE orders SET order_status = $1, accrual = $2 WHERE order_number = $3"

	AddBalance string = "UPDATE users SET balance = balance + $1 WHERE user_id = $2"

	Withdrawals string = "SELECT order_number, amount, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC;"

	GetBalance string = "SELECT balance FROM users WHERE user_id = $1;"

	WithdrawFromBalance string = "UPDATE users SET balance = balance - $1 WHERE user_id = $2"

	AddWithdraw string = "INSERT INTO withdrawals (user_id, order_number, amount) VALUES ($1, $2, $3);"

	WithdrawalSum string = "SELECT COALESCE(SUM(amount), 0) AS withdrawal_sum FROM withdrawals WHERE user_id = $1"
)
