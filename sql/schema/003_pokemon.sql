-- +goose Up
CREATE TABLE pokemon(
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    sprite TEXT NOT NULL,
    type TEXT NOT NULL,
    url TEXT NOT NULL 
);

-- +goose down
DROP TABLE pokemon;