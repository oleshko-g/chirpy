-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS chirps ADD COLUMN deleted_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS chirps DROP COLUMN deleted_at TIMESTAMPTZ;
-- +goose StatementEnd
