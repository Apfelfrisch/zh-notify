-- name: GetEvent :one
SELECT * FROM events WHERE id = ? LIMIT 1;

-- name: GetEventsForPeriod :many
SELECT * FROM events WHERE reported_at_upcoming IS NULL AND (DATE(date) >= ? and DATE(date) <= ?) ORDER BY date;

-- name: GetFreshEvents :many
SELECT * FROM events WHERE reported_at_new IS NULL ORDER BY date;

-- name: GetNakedEvents :many
SELECT * FROM events WHERE reported_at_upcoming IS NULL AND (
    artist IS NULL
    OR category IS NULL
    OR artist_url IS NULL
    OR artist_img_url IS NULL
) ORDER BY date;

-- name: MarkFreshEventsAsReported :exec
UPDATE events SET reported_at_new = ? WHERE id = ?;

-- name: MarkUpcomingEventsAsReported :exec
UPDATE events SET reported_at_upcoming = ? WHERE id = ?;

-- name: AddMetaData :exec
UPDATE events SET artist = ?, category = ?, artist_url = ?, artist_img_url = ? WHERE id = ?;

-- name: CreateEvent :exec
INSERT INTO events (name, place, status, link, date) VALUES (?, ?, ?, ?, ?)
ON CONFLICT(link) DO UPDATE SET
    name = excluded.name,
    place = excluded.place,
    status = excluded.status,
    date = excluded.date;
    -- I dont think we should update this fields on conflict - will see
    -- artist = excluded.artist,
    -- category = excluded.category,
    -- artist_url = excluded.artist_url,
    -- artist_img_url = excluded.artist_img_url;
