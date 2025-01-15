-- name: GetAllFeeds :many
SELECT * FROM feeds;

-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id) VALUES ($1, $2, $3) RETURNING *;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedByUserId :many
SELECT * FROM feeds WHERE user_id = $1;