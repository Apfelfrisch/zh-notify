package cmd

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/notify"
)

const DBProvider = "sqlite3"
const DBConnection = ":memory:"

func prepareConnection() *sql.DB {
	conn, _ := sql.Open(DBProvider, ":memory:")
	data, _ := os.ReadFile("../db/schema.sql")
	conn.Exec(string(data))

	return conn
}

func TestSaveEvents(t *testing.T) {
	t.Run("write new event", func(t *testing.T) {
		queries := db.New(prepareConnection())
		events := []notify.Event{
			{Name: "event-1", Place: "place-1", Status: "available-1", Link: "link-1", Date: time.Now()},
		}

		saveEvents(context.Background(), queries, events)

		event, err := queries.GetEvent(context.Background(), 1)
		assert.Nil(t, err)

		_, err = queries.GetEvent(context.Background(), 2)
		assert.NotNil(t, err)

		assert.Equal(t, events[0].Name, event.Name)
		assert.Equal(t, events[0].Place, event.Place)
		assert.Equal(t, events[0].Status, event.Status)
		assert.Equal(t, events[0].Link, event.Link)
		assert.Equal(t, events[0].Date.Format("02.01.06"), event.Date.Format("02.01.06"))
		assert.False(t, event.ReportedAtNew.Valid)
		assert.False(t, event.ReportedAtUpcoming.Valid)
	})

	t.Run("update an event", func(t *testing.T) {
		queries := db.New(prepareConnection())
		saveEvents(context.Background(), queries, []notify.Event{
			{Name: "event-1", Place: "place-1", Status: "available-1", Link: "link-1", Date: time.Now()},
		})
		markAllAsReported(queries)

		saveEvents(context.Background(), queries, []notify.Event{
			{Name: "event-2", Place: "place-2", Status: "available-2", Link: "link-1", Date: time.Now()},
		})

		event, err := queries.GetEvent(context.Background(), 1)
		assert.Nil(t, err)
		_, err = queries.GetEvent(context.Background(), 2)
		assert.NotNil(t, err)

		assert.NotNil(t, event)
		assert.NotNil(t, err)
		assert.Equal(t, "event-2", event.Name)
		assert.Equal(t, "place-2", event.Place)
		assert.Equal(t, "available-2", event.Status)
		assert.Equal(t, "link-1", event.Link)
		assert.True(t, event.ReportedAtNew.Valid)
		assert.True(t, event.ReportedAtUpcoming.Valid)
	})
}

func markAllAsReported(queries *db.Queries) {
	events, _ := queries.GetFreshEvents(context.Background())

	for _, event := range events {
		queries.MarkFreshEventsAsReported(context.Background(), db.MarkFreshEventsAsReportedParams{
			ReportedAtNew: sql.NullTime{Time: time.Now(), Valid: true},
			ID:            event.ID,
		})
		queries.MarkUpcomingEventsAsReported(context.Background(), db.MarkUpcomingEventsAsReportedParams{
			ReportedAtUpcoming: sql.NullTime{Time: time.Now(), Valid: true},
			ID:                 event.ID,
		})
	}
}
