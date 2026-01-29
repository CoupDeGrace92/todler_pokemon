-- name: AddPokemonToTeam :exec
INSERT INTO teams(created_at, updated_at, user_name, poke, count)
VALUES(
    NOW(),
    NOW(),
    $1,
    $2,
    $3
)
ON CONFLICT (user_name, poke)
DO UPDATE SET
    updated_at = EXCLUDED.updated_at,
    count = teams.count + $3;