package db

import (
	"context"
	"database/sql"
	"time"
)

type EventRepository interface {
	GetById(ctx context.Context, id int64) (Event, error)
	GetByLink(ctx context.Context, link string) (Event, error)
	GetFreshEvents(ctx context.Context) ([]Event, error)
	GetNakedEvents(ctx context.Context) ([]Event, error)
	GetUpcomingEvents(ctx context.Context, fromDate time.Time, daysAhead int) ([]Event, error)
	Save(ctx context.Context, event Event) error
}

func NewEventRepoFromConn(conn *sql.DB) *EventRepo {
	return &EventRepo{Queries: New(conn)}
}

func NewDbEventRepo() (*EventRepo, error) {
	queries, err := NewQueries()

	if err != nil {
		return nil, err
	}

	return &EventRepo{queries}, nil
}

type EventRepo struct {
	Queries *Queries
}

func (er *EventRepo) GetById(ctx context.Context, id int64) (Event, error) {
	return er.Queries.GetEvent(ctx, id)
}

func (er *EventRepo) GetByLink(ctx context.Context, link string) (Event, error) {
	return er.Queries.GetEventByLink(ctx, link)
}

func (er *EventRepo) GetFreshEvents(ctx context.Context) ([]Event, error) {
	return er.Queries.GetFreshEvents(ctx)
}

func (er *EventRepo) GetNakedEvents(ctx context.Context) ([]Event, error) {
	return er.Queries.GetNakedEvents(ctx)
}

func (er *EventRepo) GetUpcomingEvents(ctx context.Context, fromDate time.Time, daysAhead int) ([]Event, error) {
	nm := fromDate.AddDate(0, 0, daysAhead)
	startOfMonth := time.Date(nm.Year(), nm.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	return er.Queries.GetEventsForPeriod(ctx, GetEventsForPeriodParams{
		Date:   fromDate,
		Date_2: endOfMonth,
	})
}

func (er *EventRepo) Save(ctx context.Context, event Event) error {
	if event.ID == 0 {
		return er.Queries.CreateEvent(ctx, CreateEventParams{
			Name:         event.Name,
			Place:        event.Place,
			Status:       event.Status,
			Link:         event.Link,
			Date:         event.Date,
			ArtistImgUrl: event.ArtistImgUrl,
		})
	}

	return er.Queries.UpdateEvent(ctx, UpdateEventParams{
		Name:               event.Name,
		Place:              event.Place,
		Status:             event.Status,
		Link:               event.Link,
		Date:               event.Date,
		Artist:             event.Artist,
		Category:           event.Category,
		ArtistUrl:          event.ArtistUrl,
		ArtistImgUrl:       event.ArtistImgUrl,
		ReportedAtNew:      event.ReportedAtNew,
		ReportedAtUpcoming: event.ReportedAtUpcoming,
		PostponedDate:      event.PostponedDate,
		ID:                 event.ID,
	})
}
