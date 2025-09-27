-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN user_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd
