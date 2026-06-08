-- sqlc/queries/messages.sql

-- name: CreateMessage :one
INSERT INTO messages (
    user_id,
    username,
    content
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLastMessages :many
SELECT *
FROM messages
ORDER BY created_at DESC
LIMIT $1;

-- name: GetMessagesByUserID :many
SELECT *
FROM messages
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

