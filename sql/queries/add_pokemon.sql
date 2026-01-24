-- name: AddPokemon :exec
INSERT INTO pokemon(id, name, sprite, type, url)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
);