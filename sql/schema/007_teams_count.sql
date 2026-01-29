-- +goose Up
ALTER TABLE teams
ADD COLUMN count INT NOT NULL DEFAULT 0;

UPDATE teams
SET count = 1
WHERE count = 0;

CREATE TABLE teams_temp (
    user_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    poke TEXT NOT NULL,
    count INT NOT NULL,
    PRIMARY KEY (user_name, poke),
    CONSTRAINT fk_user FOREIGN KEY (user_name) REFERENCES users(user_name) ON DELETE CASCADE,
    CONSTRAINT fk_poke FOREIGN KEY (poke) REFERENCES pokemon(name) ON DELETE CASCADE
);

INSERT INTO teams_temp (user_name, poke, count, created_at, updated_at)
SELECT user_name, poke, SUM(count) AS count, MIN(created_at) AS created_at, MAX(updated_at) AS updated_at
FROM teams
GROUP BY user_name, poke;

DROP TABLE teams;

ALTER TABLE teams_temp RENAME TO teams;


-- +goose Down
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE teams_old (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_name TEXT NOT NULL,
    poke TEXT NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_name) REFERENCES users(user_name) ON DELETE CASCADE,
    CONSTRAINT fk_poke FOREIGN KEY (poke) REFERENCES pokemon(name) ON DELETE CASCADE
);
INSERT INTO teams_old(id, created_at, updated_at, user_name, poke)
SELECT gen_random_uuid() AS id,
       created_at,
       updated_at,
       user_name,
       poke
FROM teams t,
    generate_series(1, t.count);

DROP TABLE teams;

ALTER TABLE teams_old RENAME TO teams;
--quick note - this down migration does not preserver the updated_at and created_at tags from pre-migration