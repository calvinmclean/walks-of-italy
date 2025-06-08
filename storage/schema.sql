CREATE TABLE IF NOT EXISTS tours (
    uuid UUID PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS latest_availabilities (
    tour_uuid UUID NOT NULL,
    recorded_at DATETIME NOT NULL,
    availability_date DATETIME NOT NULL,
    raw_data TEXT NOT NULL,
    FOREIGN KEY (tour_uuid) REFERENCES tours (uuid)
);

-- add new column to track the API URL
ALTER TABLE tours
ADD COLUMN api_url TEXT NOT NULL DEFAULT '';

-- rename url to link since it is a link to the site, but not used by the application
ALTER TABLE tours
RENAME COLUMN url TO link;
