-- +goose Up
CREATE TABLE refresh_tokens(
    token TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_name TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    FOREIGN KEY (user_name) REFERENCES users(user_name) ON DELETE CASCADE
);

-- +goose down
DROP TABLE refresh_tokens;