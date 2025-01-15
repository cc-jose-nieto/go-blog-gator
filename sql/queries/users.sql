-- name: CreateUser :one
INSERT INTO users (name)
VALUES(
$1
)
RETURNING *;

-- name: GetUserByName :one
SELECT id, created_at, updated_at, name FROM users WHERE name = $1;

-- name: GetUserByID :one
SELECT id, created_at, updated_at, name FROM users WHERE id = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetAllUsers :many
SELECT id, created_at, updated_at, name FROM users;