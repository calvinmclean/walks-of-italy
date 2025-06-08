-- name: GetLatestAvailability :one
SELECT
    tour_uuid,
    recorded_at,
    availability_date,
    raw_data
FROM
    latest_availabilities
WHERE
    tour_uuid = ?
ORDER BY
    availability_date DESC
LIMIT
    1;

-- name: GetAllLatestAvailabilities :many
SELECT
    t.name,
    t.link,
    t.api_url,
    t.uuid,
    la.recorded_at,
    la.availability_date,
    la.raw_data
FROM
    latest_availabilities la
    JOIN tours t ON t.uuid = la.tour_uuid
WHERE
    la.recorded_at = (
        SELECT
            MAX(recorded_at)
        FROM
            latest_availabilities
        WHERE
            tour_uuid = t.uuid
    );

-- name: AddLatestAvailability :exec
INSERT INTO
    latest_availabilities (
        tour_uuid,
        recorded_at,
        availability_date,
        raw_data
    )
VALUES
    (?, CURRENT_TIMESTAMP, ?, ?);
