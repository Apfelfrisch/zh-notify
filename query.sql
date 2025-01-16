-- name: GetEvent :one
SELECT * FROM events WHERE id = ? LIMIT 1;

-- name: GetEventByLink :one
SELECT * FROM events WHERE link = ? LIMIT 1;

-- name: GetEventsForPeriod :many
SELECT * FROM events
    WHERE reported_at_upcoming IS NULL
    AND (
        DATE(date) >= DATE(?) AND DATE(date) <= DATE(?)
    )
ORDER BY date;

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

-- name: UpdateEvent :exec
UPDATE events
SET
    name = ?,
    place = ?,
    status = ?,
    link = ?,
    date = ?,
    artist = ?,
    reported_at_new = ?,
    reported_at_upcoming = ?,
    category = ?,
    artist_url = ?,
    artist_img_url = ?,
    postponed_date = ?
WHERE id = ?;

-- name: CreateEvent :exec
INSERT INTO events (name, place, status, link, date, artist_img_url) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(link) DO UPDATE SET
    name = excluded.name,
    place = excluded.place,
    status = excluded.status,
    date = excluded.date;
    -- I dont think we should update this fields on conflict - will see
    -- artist = excluded.artist,
    -- category = excluded.category,
    -- artist_url = excluded.artist_url,
