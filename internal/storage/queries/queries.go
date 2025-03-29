package queries

const (
	InsertUser string = "INSERT INTO users (user_id, user_name, password_hash) VALUES ($1, $2, $3);"
)
