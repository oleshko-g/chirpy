-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN is_chirpy_red bool NOT NULL DEFAULT FALSE;
COMMENT ON COLUMN users.is_chirpy_red is 'Chirpy premium subscription';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table users
DROP COLUMN is_chirpy_red;
-- +goose StatementEnd
