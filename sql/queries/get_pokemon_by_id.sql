-- name: GetPokeByID :one
SELECT * FROM pokemon
WHERE id = $1;