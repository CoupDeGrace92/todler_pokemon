-- +goose Up
CREATE TABLE teams(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_name TEXT NOT NULL,
    poke TEXT NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_name) REFERENCES users(user_name) ON DELETE CASCADE,
    CONSTRAINT fk_poke FOREIGN KEY (poke) REFERENCES pokemon(name) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE teams