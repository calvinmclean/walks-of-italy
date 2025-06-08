-- name: GetTour :one
SELECT
    *
FROM
    tours
WHERE
    uuid = ?
LIMIT
    1;

-- name: ListTours :many
SELECT
    *
FROM
    tours;

-- name: UpsertTour :exec
INSERT INTO
    tours (uuid, name, link, api_url)
VALUES
    (?, ?, ?, ?) ON CONFLICT (uuid) DO
UPDATE
SET
    name = EXCLUDED.name,
    link = EXCLUDED.link,
    api_url = EXCLUDED.api_url;

-- name: DeleteTour :exec
DELETE FROM tours
WHERE
    uuid = ?;
