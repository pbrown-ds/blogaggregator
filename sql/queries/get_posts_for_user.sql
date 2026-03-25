-- name: GetPostsForUser :many
SELECT
    posts.*,
    feeds.name AS feed_name
FROM posts
INNER JOIN feeds
    ON posts.feed_id = feeds.id
INNER JOIN feed_follows
    ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1
ORDER BY posts.created_at
LIMIT $2;