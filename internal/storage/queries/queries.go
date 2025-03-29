package queries

const (
	InsertUser string = "INSERT INTO users (user_id, user_name, password_hash) VALUES ($1, $2, $3);"

	GetUserID string = "SELECT user_id, password_hash from users WHERE user_name = $1"
)
