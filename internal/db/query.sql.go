// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const addMetaData = `-- name: AddMetaData :exec
UPDATE events SET artist = ?, category = ?, artist_url = ?, artist_img_url = ? WHERE id = ?
`

type AddMetaDataParams struct {
	Artist       sql.NullString
	Category     sql.NullString
	ArtistUrl    sql.NullString
	ArtistImgUrl sql.NullString
	ID           int64
}

func (q *Queries) AddMetaData(ctx context.Context, arg AddMetaDataParams) error {
	_, err := q.db.ExecContext(ctx, addMetaData,
		arg.Artist,
		arg.Category,
		arg.ArtistUrl,
		arg.ArtistImgUrl,
		arg.ID,
	)
	return err
}

const createEvent = `-- name: CreateEvent :exec
INSERT INTO events (name, place, status, link, date, artist_img_url) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(link) DO UPDATE SET
    name = excluded.name,
    place = excluded.place,
    status = excluded.status,
    date = excluded.date
`

type CreateEventParams struct {
	Name         string
	Place        string
	Status       string
	Link         string
	Date         time.Time
	ArtistImgUrl sql.NullString
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) error {
	_, err := q.db.ExecContext(ctx, createEvent,
		arg.Name,
		arg.Place,
		arg.Status,
		arg.Link,
		arg.Date,
		arg.ArtistImgUrl,
	)
	return err
}

const getEvent = `-- name: GetEvent :one
SELECT id, name, place, status, link, date, artist, category, artist_url, artist_img_url, reported_at_new, reported_at_upcoming, postponed_date, created_at FROM events WHERE id = ? LIMIT 1
`

func (q *Queries) GetEvent(ctx context.Context, id int64) (Event, error) {
	row := q.db.QueryRowContext(ctx, getEvent, id)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Place,
		&i.Status,
		&i.Link,
		&i.Date,
		&i.Artist,
		&i.Category,
		&i.ArtistUrl,
		&i.ArtistImgUrl,
		&i.ReportedAtNew,
		&i.ReportedAtUpcoming,
		&i.PostponedDate,
		&i.CreatedAt,
	)
	return i, err
}

const getEventByLink = `-- name: GetEventByLink :one
SELECT id, name, place, status, link, date, artist, category, artist_url, artist_img_url, reported_at_new, reported_at_upcoming, postponed_date, created_at FROM events WHERE link = ? LIMIT 1
`

func (q *Queries) GetEventByLink(ctx context.Context, link string) (Event, error) {
	row := q.db.QueryRowContext(ctx, getEventByLink, link)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Place,
		&i.Status,
		&i.Link,
		&i.Date,
		&i.Artist,
		&i.Category,
		&i.ArtistUrl,
		&i.ArtistImgUrl,
		&i.ReportedAtNew,
		&i.ReportedAtUpcoming,
		&i.PostponedDate,
		&i.CreatedAt,
	)
	return i, err
}

const getEventsForPeriod = `-- name: GetEventsForPeriod :many
SELECT id, name, place, status, link, date, artist, category, artist_url, artist_img_url, reported_at_new, reported_at_upcoming, postponed_date, created_at FROM events
    WHERE reported_at_upcoming IS NULL
    AND (
        DATE(date) >= DATE(?) AND DATE(date) <= DATE(?)
    )
ORDER BY date
`

type GetEventsForPeriodParams struct {
	Date   interface{}
	Date_2 interface{}
}

func (q *Queries) GetEventsForPeriod(ctx context.Context, arg GetEventsForPeriodParams) ([]Event, error) {
	rows, err := q.db.QueryContext(ctx, getEventsForPeriod, arg.Date, arg.Date_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Place,
			&i.Status,
			&i.Link,
			&i.Date,
			&i.Artist,
			&i.Category,
			&i.ArtistUrl,
			&i.ArtistImgUrl,
			&i.ReportedAtNew,
			&i.ReportedAtUpcoming,
			&i.PostponedDate,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFreshEvents = `-- name: GetFreshEvents :many
SELECT id, name, place, status, link, date, artist, category, artist_url, artist_img_url, reported_at_new, reported_at_upcoming, postponed_date, created_at FROM events WHERE reported_at_new IS NULL ORDER BY date
`

func (q *Queries) GetFreshEvents(ctx context.Context) ([]Event, error) {
	rows, err := q.db.QueryContext(ctx, getFreshEvents)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Place,
			&i.Status,
			&i.Link,
			&i.Date,
			&i.Artist,
			&i.Category,
			&i.ArtistUrl,
			&i.ArtistImgUrl,
			&i.ReportedAtNew,
			&i.ReportedAtUpcoming,
			&i.PostponedDate,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNakedEvents = `-- name: GetNakedEvents :many
SELECT id, name, place, status, link, date, artist, category, artist_url, artist_img_url, reported_at_new, reported_at_upcoming, postponed_date, created_at FROM events WHERE reported_at_upcoming IS NULL AND (
    artist IS NULL
    OR category IS NULL
    OR artist_url IS NULL
    OR artist_img_url IS NULL
) ORDER BY date
`

func (q *Queries) GetNakedEvents(ctx context.Context) ([]Event, error) {
	rows, err := q.db.QueryContext(ctx, getNakedEvents)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Place,
			&i.Status,
			&i.Link,
			&i.Date,
			&i.Artist,
			&i.Category,
			&i.ArtistUrl,
			&i.ArtistImgUrl,
			&i.ReportedAtNew,
			&i.ReportedAtUpcoming,
			&i.PostponedDate,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markFreshEventsAsReported = `-- name: MarkFreshEventsAsReported :exec
UPDATE events SET reported_at_new = ? WHERE id = ?
`

type MarkFreshEventsAsReportedParams struct {
	ReportedAtNew sql.NullTime
	ID            int64
}

func (q *Queries) MarkFreshEventsAsReported(ctx context.Context, arg MarkFreshEventsAsReportedParams) error {
	_, err := q.db.ExecContext(ctx, markFreshEventsAsReported, arg.ReportedAtNew, arg.ID)
	return err
}

const markUpcomingEventsAsReported = `-- name: MarkUpcomingEventsAsReported :exec
UPDATE events SET reported_at_upcoming = ? WHERE id = ?
`

type MarkUpcomingEventsAsReportedParams struct {
	ReportedAtUpcoming sql.NullTime
	ID                 int64
}

func (q *Queries) MarkUpcomingEventsAsReported(ctx context.Context, arg MarkUpcomingEventsAsReportedParams) error {
	_, err := q.db.ExecContext(ctx, markUpcomingEventsAsReported, arg.ReportedAtUpcoming, arg.ID)
	return err
}

const updateEvent = `-- name: UpdateEvent :exec
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
WHERE id = ?
`

type UpdateEventParams struct {
	Name               string
	Place              string
	Status             string
	Link               string
	Date               time.Time
	Artist             sql.NullString
	ReportedAtNew      sql.NullTime
	ReportedAtUpcoming sql.NullTime
	Category           sql.NullString
	ArtistUrl          sql.NullString
	ArtistImgUrl       sql.NullString
	PostponedDate      sql.NullTime
	ID                 int64
}

func (q *Queries) UpdateEvent(ctx context.Context, arg UpdateEventParams) error {
	_, err := q.db.ExecContext(ctx, updateEvent,
		arg.Name,
		arg.Place,
		arg.Status,
		arg.Link,
		arg.Date,
		arg.Artist,
		arg.ReportedAtNew,
		arg.ReportedAtUpcoming,
		arg.Category,
		arg.ArtistUrl,
		arg.ArtistImgUrl,
		arg.PostponedDate,
		arg.ID,
	)
	return err
}
