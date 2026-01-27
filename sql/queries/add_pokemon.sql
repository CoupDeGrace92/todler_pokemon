-- name: AddPokemon :exec
INSERT INTO pokemon(id, name, sprite, type, url, base_xp)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
);