-- name: GetTeam :many
SELECT * FROM teams
WHERE user_name = $1;