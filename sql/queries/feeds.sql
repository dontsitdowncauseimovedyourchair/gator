-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1,
        $2,
        $3,
        $4,
        $5,
        $6
       )
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedCreatorName :one
SELECT name FROM users
WHERE id = $1;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE feeds.url = $1;