package notify

import (
	"context"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
)

type EventRepository interface {
	GetById(ctx context.Context, id int64) (db.Event, error)
	GetByLink(ctx context.Context, link string) (db.Event, error)
	GetFreshEvents(ctx context.Context) ([]db.Event, error)
	GetNakedEvents(ctx context.Context) ([]db.Event, error)
	GetUpcomingEvents(ctx context.Context, fromDate time.Time, daysAhead int) ([]db.Event, error)
	Save(ctx context.Context, event db.Event) error
}

func NewDbEventRepo(queries *db.Queries) *dbEventRepo {
	repo := dbEventRepo{queries}
	return &repo
}

type dbEventRepo struct {
	queries *db.Queries
}

func (er *dbEventRepo) GetById(ctx context.Context, id int64) (db.Event, error) {
	return er.queries.GetEvent(ctx, id)
}

func (er *dbEventRepo) GetByLink(ctx context.Context, link string) (db.Event, error) {
	return er.queries.GetEventByLink(ctx, link)
}

func (er *dbEventRepo) GetFreshEvents(ctx context.Context) ([]db.Event, error) {
	return er.queries.GetFreshEvents(ctx)
}

func (er *dbEventRepo) GetNakedEvents(ctx context.Context) ([]db.Event, error) {
	return er.queries.GetNakedEvents(ctx)
}

func (er *dbEventRepo) GetUpcomingEvents(ctx context.Context, fromDate time.Time, daysAhead int) ([]db.Event, error) {
	nm := fromDate.AddDate(0, 0, daysAhead)
	startOfMonth := time.Date(nm.Year(), nm.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	return er.queries.GetEventsForPeriod(ctx, db.GetEventsForPeriodParams{
		Date:   fromDate,
		Date_2: endOfMonth,
	})
}

func (er *dbEventRepo) Save(ctx context.Context, event db.Event) error {
	if event.ID == 0 {
		return er.queries.CreateEvent(ctx, db.CreateEventParams{
			Name:   event.Name,
			Place:  event.Place,
			Status: event.Status,
			Link:   event.Link,
			Date:   time.Now(),
		})
	}

	return er.queries.UpdateEvent(ctx, db.UpdateEventParams{
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
