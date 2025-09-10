-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'urls'
    ) THEN
        CREATE TABLE urls (
            id SERIAL PRIMARY KEY,
            original_url VARCHAR(255) NOT NULL,
            short_url VARCHAR(255) NOT NULL
        );
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls; 
-- +goose StatementEnd
