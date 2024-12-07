package notify

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
)

func TestSendMonthEvents(t *testing.T) {
	var tests = []struct {
		name      string
		wantCount int
		events    []db.Event
	}{
		{
			"one match, the other is after range",
			1,
			[]db.Event{
				{
					Date: time.Now().AddDate(0, 1, -time.Now().Day()+1).Add(time.Hour * 24 * 5), // first of month + 5 days
					Name: "Event 1",
				},
				{
					Date: time.Now().AddDate(0, 1, -time.Now().Day()+1).Add(time.Hour * 24 * 40), // next month + 40 days
					Name: "Event 2",
				},
			},
		},
		{
			"one match, the other is before range",
			1,
			[]db.Event{
				{
					Date: time.Now().AddDate(0, 0, -time.Now().Day()+1).Add(time.Hour * 24 * 5), // first of this month
					Name: "Event 1",
				},
				{
					Date: time.Now().AddDate(0, 1, -time.Now().Day()+1).Add(time.Hour * 24 * 27), // first day of next month + 27 days
					Name: "Event 2",
				},
			},
		},
		{
			"two matches",
			2,
			[]db.Event{
				{
					Date: time.Now().AddDate(0, 1, -time.Now().Day()+1).Add(time.Hour * 24 * 5), // first day of next month
					Name: "Event 1",
				},
				{
					Date: time.Now().AddDate(0, 1, -time.Now().Day()+1).Add(time.Hour * 24 * 26), // first day of next month + 27 days
					Name: "Event 2",
				},
			},
		},
		{
			"no matches, both before range",
			0,
			[]db.Event{
				{
					Date: time.Now().AddDate(0, 0, -time.Now().Day()+1).Add(time.Hour * 24 * 5), // first day of next month
					Name: "Event 1",
				},
				{
					Date: time.Now().AddDate(0, 0, -time.Now().Day()+1).Add(time.Hour * 24 * 26), // first day of next month + 27 days
					Name: "Event 2",
				},
			},
		},
		{
			"no matches, both after range",
			0,
			[]db.Event{
				{
					Date: time.Now().AddDate(0, 0, -time.Now().Day()+1).Add(time.Hour * 24 * 5), // first day of next month
					Name: "Event 1",
				},
				{
					Date: time.Now().AddDate(0, 0, -time.Now().Day()+1).Add(time.Hour * 24 * 26), // first day of next month + 27 days
					Name: "Event 2",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := InMemoryEventDriver{}
			repo := InMemoryEventRepo{events: test.events}
			notificator := notificator{queries: repo, sender: &driver}

			notificator.SendMonthlyEvents(context.Background(), "receiver")

			countMatches := strings.Count(driver.message, "Event")

			if countMatches != test.wantCount {
				t.Errorf("got %v message(s), want %v", countMatches, test.wantCount)
			}
		})
	}
}

type InMemoryEventRepo struct {
	events []db.Event
}

func (er InMemoryEventRepo) CreateEvent(ctx context.Context, arg db.CreateEventParams) error {
	return errors.New("unimplemented")
}

func (er InMemoryEventRepo) GetEvent(ctx context.Context, id int64) (db.Event, error) {
	return db.Event{}, errors.New("unimplemented")
}

func (er InMemoryEventRepo) GetEventsForPeriod(ctx context.Context, arg db.GetEventsForPeriodParams) ([]db.Event, error) {
	var matchedEvents []db.Event

	for _, event := range er.events {
		if event.Date.Before(arg.Date) {
			continue
		}

		if event.Date.After(arg.Date_2) {
			continue
		}

		matchedEvents = append(matchedEvents, event)
	}

	return matchedEvents, nil
}

func (er InMemoryEventRepo) MarkFreshEventsAsReported(ctx context.Context, arg db.MarkFreshEventsAsReportedParams) error {
	return errors.New("Test ")
}

func (er InMemoryEventRepo) GetFreshEvents(ctx context.Context) ([]db.Event, error) {
	return nil, errors.New("Test")
}

type InMemoryEventDriver struct {
	message string
}

func (d *InMemoryEventDriver) Send(ctx context.Context, receiver string, message string) error {
	d.message = message
	return nil
}

func (d *InMemoryEventDriver) SendImage(ctx context.Context, receiver string, message string, image []byte, mimeType string) error {
	return nil
}
