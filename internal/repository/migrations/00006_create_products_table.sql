-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type_id INT NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    reception_id UUID NOT NULL REFERENCES receptions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_products_reception_id_date_time_desc 
    ON products(reception_id, date_time DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
