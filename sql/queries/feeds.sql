-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;



-- name: GetFeeds :many
Select f.name, f.url, u.name
FROM feeds f
INNER JOIN users u 
ON u.id = f.user_id;



-- name: GetFeed :one
SELECT id FROM feeds WHERE url = $1;


-- name: MarkFeedFetched :exec
UPDATE feeds SET
updated_at = CURRENT_TIMESTAMP, last_fetched_at = CURRENT_TIMESTAMP
WHERE id = $1;



-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;
