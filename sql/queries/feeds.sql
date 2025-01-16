-- name: GetAllFeeds :many
SELECT * FROM feeds;

-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id) VALUES ($1, $2, $3) RETURNING *;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedByUserId :many
SELECT * FROM feeds WHERE user_id = $1;

-- name: UpdateFeedLastFetchedAt :exec
UPDATE feeds SET last_fetched_at = now(), updated_at = now() WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 1;

-- name: GetAllFeedToFetch :many
SELECT * FROM feeds ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 5;