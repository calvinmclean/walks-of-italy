CREATE TABLE IF NOT EXISTS latest_availabilities (
    tour_uuid UUID NOT NULL,
    recorded_at DATETIME NOT NULL,
    availability_date DATETIME NOT NULL,
    raw_data TEXT NOT NULL
);
