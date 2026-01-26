-- +goose Up
ALTER TABLE pokemon
ADD CONSTRAINT unique_name UNIQUE (name);

-- +goose Down
ALTER TABLE pokemon
DROP CONSTRAINT unique_name;