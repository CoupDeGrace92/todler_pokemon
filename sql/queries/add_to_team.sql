-- name: AddPokemonToTeam :exec
INSERT INTO teams(id, created_at, updated_at, user_name, poke)
VALUES(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
);