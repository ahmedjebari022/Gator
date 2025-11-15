-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_followers (id, created_at, updated_at, user_id, feed_id)
    VALUES(
        $1,
        $2,
        $3,
        $4,
        $5
    )RETURNING *
)
SELECT inserted_feed_follow.*,
        feeds.name AS feed_name,
        users.name AS users_name
FROM inserted_feed_follow
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users ON users.id = inserted_feed_follow.user_id;


-- name: GetFeedFollowsForUser :many
SELECT f.name 
FROM feeds f
INNER JOIN feed_followers e ON f.id = e.feed_id
WHERE e.user_id = $1;


-- name: DeleteFeedFollows :exec
DELETE FROM feed_followers WHERE user_id = $1 and feed_id = $2;