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
	eventTmpl := notify.CrawledEvent{Name: "event-1", Place: "place-1", Status: "available-1", Link: "link-1", Date: time.Now()}

	t.Run("write new event", func(t *testing.T) {
		repo := notify.NewDbEventRepo(db.New(prepareConnection()))
		events := []notify.CrawledEvent{eventTmpl}

		saveEvents(context.Background(), repo, events)

		event, err := repo.GetById(context.Background(), 1)
		assert.Nil(t, err)

		_, err = repo.GetById(context.Background(), 2)
		assert.NotNil(t, err)

		assert.Equal(t, eventTmpl.Name, event.Name)
		assert.Equal(t, eventTmpl.Place, event.Place)
		assert.Equal(t, eventTmpl.Status, event.Status)
		assert.Equal(t, eventTmpl.Link, event.Link)
		assert.Equal(t, eventTmpl.Date.Format("02.01.06"), event.Date.Format("02.01.06"))
		assert.False(t, event.ReportedAtNew.Valid)
		assert.False(t, event.ReportedAtUpcoming.Valid)
	})

	t.Run("update an event", func(t *testing.T) {
		repo := notify.NewDbEventRepo(db.New(prepareConnection()))
		saveEvents(context.Background(), repo, []notify.CrawledEvent{eventTmpl})
		markAllAsReported(repo)

		saveEvents(context.Background(), repo, []notify.CrawledEvent{
			{Name: "event-2", Place: "place-2", Status: "available-2", Link: "link-1", Date: time.Now()},
		})

		event, err := repo.GetById(context.Background(), 1)
		assert.Nil(t, err)
		_, err = repo.GetById(context.Background(), 2)
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

	t.Run("update a prosponed event", func(t *testing.T) {
		repo := notify.NewDbEventRepo(db.New(prepareConnection()))
		saveEvents(context.Background(), repo, []notify.CrawledEvent{eventTmpl})
		markAllAsReported(repo)

		saveEvents(context.Background(), repo, []notify.CrawledEvent{
			{Name: "event-2", Place: "place-2", Status: "available-2", Link: "link-1", Date: time.Now().AddDate(0, 1, 0)},
		})

		event, err := repo.GetById(context.Background(), 1)
		assert.Nil(t, err)
		_, err = repo.GetById(context.Background(), 2)
		assert.NotNil(t, err)

		assert.NotNil(t, event)
		assert.NotNil(t, err)
		assert.Equal(t, "event-2", event.Name)
		assert.Equal(t, "place-2", event.Place)
		assert.Equal(t, "available-2", event.Status)
		assert.Equal(t, "link-1", event.Link)
		assert.False(t, event.ReportedAtNew.Valid)
		assert.False(t, event.ReportedAtUpcoming.Valid)
		assert.True(t, event.PostponedDate.Valid)
	})
}

func markAllAsReported(repo notify.EventRepository) {
	events, _ := repo.GetFreshEvents(context.Background())

	for _, event := range events {
		event.ReportedAtNew = sql.NullTime{Time: time.Now(), Valid: true}
		event.ReportedAtUpcoming = sql.NullTime{Time: time.Now(), Valid: true}

		repo.Save(context.Background(), event)
	}
}
