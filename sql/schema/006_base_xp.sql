-- +goose Up
ALTER TABLE pokemon
Add base_xp INT NOT NULL DEFAULT 0;


-- +goose Down
ALTER TABLE pokemon
DROP COLUMN base_xp;