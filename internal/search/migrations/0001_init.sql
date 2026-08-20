CREATE TABLE IF NOT EXISTS documents (
    post_id          text PRIMARY KEY,
    author_id        text NOT NULL,
    author_username  text NOT NULL,
    content          text NOT NULL,
    removed          boolean NOT NULL DEFAULT false,
    author_suspended boolean NOT NULL DEFAULT false,
    created_at       timestamptz NOT NULL,
    updated_at       timestamptz NOT NULL DEFAULT now(),
    search_vector    tsvector GENERATED ALWAYS AS (to_tsvector('english', content)) STORED
);

CREATE INDEX IF NOT EXISTS documents_search_vector_idx ON documents USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS documents_author_id_idx ON documents (author_id);
