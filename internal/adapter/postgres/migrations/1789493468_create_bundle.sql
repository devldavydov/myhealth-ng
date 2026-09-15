-- +goose Up
CREATE TABLE IF NOT EXISTS bundle (
    key  TEXT NOT NULL,
    name TEXT NOT NULL,
    PRIMARY KEY (key)
);

CREATE TABLE IF NOT EXISTS bundle_item (
    bundle_key TEXT NOT NULL,
    food_key   TEXT NOT NULL,
    weight     REAL NOT NULL CHECK (weight > 0),
    PRIMARY KEY (bundle_key, food_key),
    FOREIGN KEY (bundle_key) REFERENCES bundle(key) ON DELETE CASCADE,
    FOREIGN KEY (food_key) REFERENCES food(key) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS bundle_name_lower_idx ON bundle (lower(name));

-- +goose Down
DROP TABLE IF EXISTS bundle_item;
DROP TABLE IF EXISTS bundle;
