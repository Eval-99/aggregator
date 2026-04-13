-- name: GetUser :one
SELECT
    name
FROM
    users
WHERE
    name = $1;
