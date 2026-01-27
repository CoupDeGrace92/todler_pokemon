-- name: ResetUserTeam :exec
DELETE FROM teams
WHERE user_name = $1;