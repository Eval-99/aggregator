-- name: AllFeeds :many
SELECT *
FROM feeds
INNER JOIN users
    ON feeds.user_id = users.id;
