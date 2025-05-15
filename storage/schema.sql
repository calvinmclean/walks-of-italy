CREATE TABLE IF NOT EXISTS latest_availabilities (
    tour_uuid UUID NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    availability_date TIMESTAMPTZ NOT NULL,
    raw_data TEXT
);
