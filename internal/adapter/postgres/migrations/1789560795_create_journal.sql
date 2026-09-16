-- +goose Up
CREATE TYPE meal_type AS ENUM (
    'завтрак',
    'до обеда',
    'обед',
    'полдник',
    'до ужина',
    'ужин'
);

CREATE TABLE IF NOT EXISTS journal (
    user_id     TEXT      NOT NULL,
    dt          DATE      NOT NULL,
    meal        meal_type NOT NULL,
    food_key    TEXT      NOT NULL,
    food_weight REAL      NOT NULL CHECK (food_weight > 0),
    PRIMARY KEY (user_id, dt, meal, food_key),
    FOREIGN KEY (food_key) REFERENCES food(key) ON DELETE RESTRICT
);

-- +goose Down
DROP TABLE IF EXISTS journal;
DROP TYPE IF EXISTS meal_type;
