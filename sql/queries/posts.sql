-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;


-- name: GetPostsForUser :many
SELECT p.title, p.url, p.description, p.published_at, p.feed_id, f.name AS feed_name
FROM posts p
INNER JOIN feeds f ON  f.id = p.feed_id
INNER JOIN feed_followers ff ON ff.user_id = f.user_id
WHERE ff.user_id = $1
ORDER BY P.published_at DESC
limit $2;