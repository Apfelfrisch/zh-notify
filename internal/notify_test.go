package internal

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/apfelfrisch/zh-notify/internal/db"
	"github.com/apfelfrisch/zh-notify/internal/transport"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestSendUpcomingEvents(t *testing.T) {
	var tests = []struct {
		name             string
		sendMessageCount int
		events           []db.Event
	}{
		{
			"one match, the others are after range",
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
			notificator := Notificator{&repo, &driver}

			notificator.SendMonthlyEvents(context.Background(), "receiver")

			sendEvents := lo.Filter(repo.events, func(event db.Event, index int) bool { return event.ReportedAtUpcoming.Valid })

			assert.Len(t, driver.message, test.sendMessageCount)
			assert.Len(t, sendEvents, test.sendMessageCount)
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
		notificator := Notificator{&repo, &driver}

		notificator.SendMonthlyEvents(context.Background(), "receiver")

		assert.Len(t, driver.message, 0)
	})

	t.Run("test upcoming event content", func(t *testing.T) {
		driver := InMemoryEventDriver{}
		event := db.Event{
			ID:        1,
			Date:      time.Now().AddDate(0, 0, NOTIFY_DAYS_AHEAD),
			Name:      "Event 1",
			Link:      "link-1-2-3",
			Place:     "my-location",
			Status:    "sold out",
			ArtistUrl: sql.NullString{String: "artist-url", Valid: true},
		}
		repo := InMemoryEventRepo{events: []db.Event{event}}
		notificator := Notificator{&repo, &driver}

		notificator.SendMonthlyEvents(context.Background(), "receiver")

		assert.Len(t, driver.message, 1)

		assert.Contains(t, driver.message[0], event.Date.Format(DATE_FORMAT))
		assert.Contains(t, driver.message[0], event.Status)
		assert.Contains(t, driver.message[0], "Location: "+event.Place)
		assert.Contains(t, driver.message[0], "Spotify: "+event.ArtistUrl.String)
		assert.Contains(t, driver.message[0], "Info: "+event.Link)
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
			notificator := Notificator{&repo, &driver}

			notificator.SendFreshEvents(context.Background(), "receiver")

			assert.Len(t, driver.message, test.sendMessageCount)
			assert.Len(
				t,
				lo.Filter(repo.events, func(event db.Event, index int) bool { return event.ReportedAtNew.Valid }),
				len(test.events),
			)
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
		notificator := Notificator{&repo, &driver}

		notificator.SendFreshEvents(context.Background(), "receiver")

		assert.Len(t, driver.message, 0)
	})

	t.Run("test fresh event content", func(t *testing.T) {
		driver := InMemoryEventDriver{}
		event := db.Event{
			ID:        1,
			Date:      time.Now().AddDate(0, 0, NOTIFY_DAYS_AHEAD),
			Name:      "Event 1",
			Link:      "link-1-2-3",
			Place:     "my-location",
			Status:    "available",
			ArtistUrl: sql.NullString{String: "artist-url", Valid: true},
		}
		repo := InMemoryEventRepo{events: []db.Event{event}}
		notificator := Notificator{&repo, &driver}

		notificator.SendFreshEvents(context.Background(), "receiver")

		assert.Len(t, driver.message, 1)

		assert.Contains(t, driver.message[0], event.Date.Format(DATE_FORMAT))
		assert.NotContains(t, driver.message[0], event.Status)
		assert.Contains(t, driver.message[0], "Location: "+event.Place)
		assert.Contains(t, driver.message[0], "Spotify: "+event.ArtistUrl.String)
		assert.Contains(t, driver.message[0], "Info: "+event.Link)
	})

	// Todo...
	// t.Run("....", func(t *testing.T) {
	// 	mimeType, image := getEventImage(db.Event{
	// 		ArtistImgUrl: sql.NullString{String: "https://www.zollhaus-leer.com/wp-content/uploads/2024/11/Kachel_Schlagzeugmafia.png", Valid: true},
	// 	})

	// 	file, _ := os.Create("test.png")
	// 	defer file.Close()
	// 	file.Write(image)

	// 	fmt.Println(mimeType)
	// })
}

type InMemoryEventDriver struct {
	message []string
}

func (d *InMemoryEventDriver) SendWithImage(arg transport.SendImageParams) error {
	d.message = append(d.message, arg.Message)
	return nil
}

type InMemoryEventRepo struct {
	events []db.Event
}

func (er *InMemoryEventRepo) GetUpcomingEvents(ctx context.Context, fromDate time.Time, daysAhead int) ([]db.Event, error) {
	return lo.Filter(er.events, func(event db.Event, index int) bool {
		nm := fromDate.AddDate(0, 0, daysAhead)
		startOfMonth := time.Date(nm.Year(), nm.Month(), 1, 0, 0, 0, 0, time.Local)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

		if event.ReportedAtUpcoming.Valid {
			return false
		}
		if event.Date.Before(fromDate) {
			return false
		}
		if event.Date.After(endOfMonth) {
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

func (er *InMemoryEventRepo) GetById(ctx context.Context, id int64) (db.Event, error) {
	return db.Event{}, errors.New("unimplemented")
}

func (er *InMemoryEventRepo) GetByLink(ctx context.Context, link string) (db.Event, error) {
	return db.Event{}, errors.New("unimplemented")
}

func (er *InMemoryEventRepo) GetNakedEvents(ctx context.Context) ([]db.Event, error) {
	return []db.Event{}, errors.New("unimplemented")
}

func (er *InMemoryEventRepo) Save(ctx context.Context, event db.Event) error {
	for i := range er.events {
		if er.events[i].ID != event.ID {
			continue
		}
		er.events[i] = event
	}
	return nil
}
