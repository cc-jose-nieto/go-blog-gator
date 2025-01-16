-- name: CreatePost :one
INSERT INTO posts (title, url, description, feed_id, published_at) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetPostsForUser :many
SELECT * FROM posts ORDER BY created_at ASC LIMIT $1;
