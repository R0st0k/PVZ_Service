-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS product_types (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO product_types(id, name) VALUES 
(1, 'электроника'),
(2, 'одежда'),
(3, 'обувь')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_types;
-- +goose StatementEnd
