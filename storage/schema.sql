CREATE TABLE IF NOT EXISTS latest_availabilities (
    tour_uuid UUID NOT NULL,
    recorded_at DATETIME NOT NULL,
    availability_date DATETIME NOT NULL,
    raw_data TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tours (
    uuid UUID PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL
);
