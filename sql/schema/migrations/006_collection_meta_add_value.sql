
-- +goose Up
ALTER TABLE collection_meta
ADD COLUMN value VARCHAR(64) NOT NULL DEFAULT ('no_value');

-- +goose Down
ALTER TABLE collection_meta DROP value;
