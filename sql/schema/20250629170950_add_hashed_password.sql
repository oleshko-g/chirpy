-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN hashed_password TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS hashed_password;
-- +goose StatementEnd
