-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN is_deleted BOOL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN IF EXISTS is_deleted;
-- +goose StatementEnd
