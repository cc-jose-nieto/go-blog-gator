-- +goose Up
CREATE TABLE feed_follows (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    CONSTRAINT feed_follows_user_id_feed_id UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;