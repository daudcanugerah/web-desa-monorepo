-- +goose Up
-- +goose StatementBegin
ALTER TABLE struktur_organisasi
    ADD COLUMN description TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE struktur_organisasi
    DROP COLUMN description;
-- +goose StatementEnd