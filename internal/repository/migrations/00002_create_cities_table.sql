-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS cities (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO cities(id, name) VALUES 
(1, 'Москва'),
(2, 'Санкт-Петербург'),
(3, 'Казань')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cities;
-- +goose StatementEnd
