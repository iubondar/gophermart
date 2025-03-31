package queries

const (
	InsertUser string = "INSERT INTO users (user_id, user_name, password_hash) VALUES ($1, $2, $3);"

	GetUserID string = "SELECT user_id, password_hash from users WHERE user_name = $1"

	GetUserIDByOrderNum string = "SELECT user_id from orders WHERE order_number = $1"

	RegisterOrder string = "INSERT INTO orders (user_id, order_number, order_status) VALUES ($1, $2, $3);"
)
