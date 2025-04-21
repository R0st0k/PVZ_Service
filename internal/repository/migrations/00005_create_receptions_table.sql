-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'reception_status') THEN
        CREATE TYPE reception_status AS ENUM ('in_progress', 'close');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pvz_id UUID NOT NULL REFERENCES pvz(id) ON DELETE CASCADE,
    status reception_status NOT NULL DEFAULT 'in_progress'
);

CREATE INDEX IF NOT EXISTS idx_receptions_in_progress_date_time_desc 
    ON receptions(date_time DESC)
    WHERE status = 'in_progress';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS receptions;

DROP TYPE IF EXISTS reception_status;
-- +goose StatementEnd
