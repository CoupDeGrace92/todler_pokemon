-- name: GetPokeByName :one
SELECT * FROM pokemon
WHERE name = $1;