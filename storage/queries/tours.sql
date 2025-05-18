-- name: GetTour :one
SELECT * FROM tours
WHERE uuid = ? LIMIT 1;

-- name: ListTours :many
SELECT * FROM tours;

-- name: UpsertTour :exec
INSERT INTO tours (
  uuid, name, url
) VALUES (
  ?, ?, ?
) ON CONFLICT (uuid)
DO UPDATE SET
  name = EXCLUDED.name,
  url = EXCLUDED.url;

-- name: DeleteTour :exec
DELETE FROM tours WHERE uuid = ?;
