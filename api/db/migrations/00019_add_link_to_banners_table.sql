-- +goose Up
-- +goose StatementBegin
ALTER TABLE banners ADD COLUMN link VARCHAR(500) DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE banners DROP COLUMN IF EXISTS link;
-- +goose StatementEnd