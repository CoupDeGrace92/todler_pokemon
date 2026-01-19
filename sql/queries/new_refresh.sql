-- name: InsertRefresh :exec
INSERT INTO refresh_tokens(token, created_at, updated_at, user_name, expires_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
);