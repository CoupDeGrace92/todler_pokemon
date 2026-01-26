-- name: ValidatePokemon :one
SELECT COUNT(*) 
FROM pokemon 
WHERE name = $1;