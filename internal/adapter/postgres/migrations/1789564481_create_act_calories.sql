-- +goose Up
CREATE TABLE IF NOT EXISTS act_calories (
    user_id TEXT NOT NULL,
    dt      DATE NOT NULL,
    value   REAL NOT NULL,
    PRIMARY KEY (user_id, dt)
);

-- +goose Down
DROP TABLE IF EXISTS act_calories;
