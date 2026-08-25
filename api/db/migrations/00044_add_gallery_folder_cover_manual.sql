-- +goose Up
-- Manual folder-cover override support (upload migration task 2.1).
-- When cover_manual is TRUE, RecomputeFolderCover leaves the cover untouched;
-- only explicit SetFolderCover (or media deletion of the cover itself) clears it.
ALTER TABLE gallery_folders
    ADD COLUMN cover_manual BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE gallery_folders
    DROP COLUMN IF EXISTS cover_manual;
