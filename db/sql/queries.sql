-- name: UpsertUser :exec
INSERT INTO users (user_id, lastfm_username)
VALUES (:user_id, :lastfm_username)
ON CONFLICT(user_id) DO UPDATE SET lastfm_username = excluded.lastfm_username;

-- name: GetUserByID :one
SELECT user_id, lastfm_username, created_at
FROM users
WHERE user_id = :user_id;

-- name: GetUserByLastFM :one
SELECT user_id, lastfm_username, created_at
FROM users
WHERE lastfm_username = :lastfm_username;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = :user_id;
