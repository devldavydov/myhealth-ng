-- +goose Up
CREATE TABLE IF NOT EXISTS sport (
    key     TEXT NOT NULL,
    name    TEXT NOT NULL,
    unit    TEXT NOT NULL,
    comment TEXT NOT NULL,
    PRIMARY KEY (key)
);

CREATE TABLE IF NOT EXISTS sport_activity (
    user_id   TEXT NOT NULL,
    dt        DATE NOT NULL,
    sport_key TEXT NOT NULL,
    sets      REAL[] NOT NULL,
    PRIMARY KEY (user_id, dt, sport_key),
    FOREIGN KEY (sport_key) REFERENCES sport(key) ON DELETE RESTRICT
);

-- +goose Down
DROP TABLE IF EXISTS sport_activity;
DROP TABLE IF EXISTS sport;
