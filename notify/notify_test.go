package notify

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/samber/lo"
)

func TestSendUpcomingEvents(t *testing.T) {
	var tests = []struct {
		name             string
		sendMessageCount int
		events           []db.Event
	}{
		{
			"one match, the other is after range",
			1,
			[]db.Event{
				{ID: 1, Date: time.Now().AddDate(0, 0, NOTIFY_DAYS_AHEAD), Name: "Event 1"},
				{ID: 2, Date: time.Now().AddDate(0, 2, 0), Name: "Event 2"},
			},
		},
		{
			"one match, the other is before range",
			1,
			[]db.Event{
				{ID: 1, Date: time.Now().AddDate(0, 0, NOTIFY_DAYS_AHEAD), Name: "Event 1"},
				{ID: 2, Date: time.Now().AddDate(0, 0, -1), Name: "Event 2"},
			},
		},
		{
			"two matches",
			2,
			[]db.Event{
				{ID: 1, Date: time.Now().AddDate(0, 0, 1), Name: "Event 1"},
				{ID: 2, Date: time.Now().AddDate(0, 0, NOTIFY_DAYS_AHEAD), Name: "Event 2"},
			},
		},
		{
			"no matches, both before/after range",
			0,
			[]db.Event{
				{ID: 1, Date: time.Now().AddDate(0, 0, -1), Name: "Event 1"},
				{ID: 2, Date: time.Now().AddDate(0, 2, 0), Name: "Event 2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := InMemoryEventDriver{}
			repo := InMemoryEventRepo{events: test.events}
			notificator := notificator{queries: &repo, sender: &driver}

			notificator.SendMonthlyEvents(context.Background(), "receiver")

			if len(driver.message) != test.sendMessageCount {
				t.Errorf("got %v send message(s), want %v", len(driver.message), test.sendMessageCount)
			}

			markedAsSend := lo.Filter(repo.events, func(event db.Event, index int) bool { return event.ReportedAtUpcoming.Valid })
			if len(markedAsSend) != test.sendMessageCount {
				t.Errorf("got %v market as send message(s), want %v", len(markedAsSend), test.sendMessageCount)
			}
		})
	}

	t.Run("no matches, both already send", func(t *testing.T) {
		driver := InMemoryEventDriver{}
		repo := InMemoryEventRepo{events: []db.Event{
			{
				ID:                 1,
				Date:               time.Now().AddDate(0, 0, NOTIFY_DAYS_AHEAD),
				ReportedAtUpcoming: sql.NullTime{Time: time.Now(), Valid: true},
				Name:               "Event 1",
			},
		}}
		notificator := notificator{queries: &repo, sender: &driver}

		notificator.SendMonthlyEvents(context.Background(), "receiver")

		if len(driver.message) != 0 {
			t.Errorf("got %v message(s), want %v", len(driver.message), 0)
		}
	})
}

func TestSendFreshEvents(t *testing.T) {
	var tests = []struct {
		name             string
		sendMessageCount int
		events           []db.Event
	}{
		{
			"one match, the other was already send",
			1,
			[]db.Event{
				{ID: 1, Date: time.Now().AddDate(0, 0, 1), Name: "Event 1"},
				{ID: 2, Date: time.Now().AddDate(0, 2, 0), ReportedAtNew: sql.NullTime{Time: time.Now(), Valid: true}, Name: "Event 2"},
			},
		},
		{
			"two matches",
			2,
			[]db.Event{
				{ID: 1, Date: time.Now().AddDate(0, 0, 1), Name: "Event 1"},
				{ID: 2, Date: time.Now().AddDate(0, 2, 0), Name: "Event 2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := InMemoryEventDriver{}
			repo := InMemoryEventRepo{events: test.events}
			notificator := notificator{queries: &repo, sender: &driver}

			notificator.SendFreshEvents(context.Background(), "receiver")

			if len(driver.message) != test.sendMessageCount {
				t.Errorf("got %v send message(s), want %v", len(driver.message), test.sendMessageCount)
			}

			markedAsSend := lo.Filter(repo.events, func(event db.Event, index int) bool { return event.ReportedAtNew.Valid })
			if len(markedAsSend) != len(test.events) {
				t.Errorf("got %v market as send message(s), want %v", len(markedAsSend), test.sendMessageCount)
			}
		})
	}

	t.Run("no matches, event has already taken palce", func(t *testing.T) {
		driver := InMemoryEventDriver{}
		repo := InMemoryEventRepo{events: []db.Event{
			{
				ID:   1,
				Date: time.Now().AddDate(0, 0, -1),
				Name: "Event 1",
			},
		}}
		notificator := notificator{queries: &repo, sender: &driver}

		notificator.SendFreshEvents(context.Background(), "receiver")

		if len(driver.message) != 0 {
			t.Errorf("got %v message(s), want %v", len(driver.message), 0)
		}
	})
}

type InMemoryEventRepo struct {
	events []db.Event
}

func (er *InMemoryEventRepo) CreateEvent(ctx context.Context, arg db.CreateEventParams) error {
	return errors.New("unimplemented")
}

func (er *InMemoryEventRepo) GetEvent(ctx context.Context, id int64) (db.Event, error) {
	return db.Event{}, errors.New("unimplemented")
}

func (er *InMemoryEventRepo) GetEventsForPeriod(ctx context.Context, arg db.GetEventsForPeriodParams) ([]db.Event, error) {
	return lo.Filter(er.events, func(event db.Event, index int) bool {
		if event.ReportedAtUpcoming.Valid {
			return false
		}
		if event.Date.Before(arg.Date) {
			return false
		}
		if event.Date.After(arg.Date_2) {
			return false
		}
		return true
	}), nil
}

func (er *InMemoryEventRepo) GetFreshEvents(ctx context.Context) ([]db.Event, error) {
	return lo.Filter(er.events, func(event db.Event, index int) bool {
		if event.Date.Before(time.Now()) {
			return false
		}
		return !event.ReportedAtNew.Valid
	}), nil
}

func (er *InMemoryEventRepo) MarkFreshEventsAsReported(ctx context.Context, arg db.MarkFreshEventsAsReportedParams) error {
	for i := range er.events {
		if er.events[i].ID != arg.ID {
			continue
		}
		er.events[i].ReportedAtNew = arg.ReportedAtNew
	}
	return nil
}

func (er *InMemoryEventRepo) MarkUpcomingEventsAsReported(ctx context.Context, arg db.MarkUpcomingEventsAsReportedParams) error {
	for i := range er.events {
		if er.events[i].ID != arg.ID {
			continue
		}
		er.events[i].ReportedAtUpcoming = arg.ReportedAtUpcoming
	}
	return nil
}

type InMemoryEventDriver struct {
	message []string
}

func (d *InMemoryEventDriver) SendWithImage(arg SendImageParams) error {
	d.message = append(d.message, arg.message)
	return nil
}
