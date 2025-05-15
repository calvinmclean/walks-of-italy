-- name: GetLatestAvailability :one
SELECT * FROM latest_availabilities
WHERE tour_uuid = ? ORDER BY availability_date DESC LIMIT 1;

-- name: AddLatestAvailability :exec
INSERT INTO latest_availabilities (
    tour_uuid,
    recorded_at,
    availability_date,
    raw_data
) VALUES (?, datetime('now'), ?, ?);
