
-- +goose Up
CREATE TABLE collection_meta (
      id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      name TEXT NOT NULL
);

-- +goose Down
DROP TABLE collection_meta;
