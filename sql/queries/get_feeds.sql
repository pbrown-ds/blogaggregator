-- name: GetFeeds :many
SELECT *
FROM feeds
ORDER BY created_at;