-- +goose Up
CREATE TABLE IF NOT EXISTS food (
    key     TEXT NOT NULL,
    name    TEXT NOT NULL,
    brand   TEXT NOT NULL,
    cal100  REAL NOT NULL,
    prot100 REAL NOT NULL,
    fat100  REAL NOT NULL,
    carb100 REAL NOT NULL,
    comment TEXT NOT NULL,
    PRIMARY KEY (key)
);

-- +goose Down
DROP TABLE IF EXISTS food;
