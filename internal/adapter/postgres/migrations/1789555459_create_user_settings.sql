-- +goose Up
CREATE TABLE IF NOT EXISTS user_settings (
    user_id                     TEXT NOT NULL,
    default_daily_calorie_limit INTEGER NOT NULL CHECK (default_daily_calorie_limit BETWEEN 1 AND 10000),
    PRIMARY KEY (user_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_settings;
