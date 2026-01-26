-- name: GetUserFromID :one
SELECT user_name FROM users
WHERE id = $1;