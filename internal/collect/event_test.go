package collect

import (
	"database/sql"
	"testing"
	"time"

	"github.com/apfelfrisch/zh-notify/internal/db"
	"github.com/apfelfrisch/zh-notify/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestMapCollectedEventToDbvents(t *testing.T) {
	tmplEvent := Event{
		Name:         "cName",
		Place:        "cPlace",
		Date:         time.Now().AddDate(0, 1, 0),
		Status:       "cStatus",
		Link:         "cLink",
		ArtistImgUrl: "cArtistUrl",
	}

	var tests = []struct {
		name           string
		collectedEvent Event
		dbEvent        db.Event
		expEvent       db.Event
	}{
		{
			"map a unsaved event",
			tmplEvent,
			db.Event{},
			db.Event{
				Name:         tmplEvent.Name,
				Place:        tmplEvent.Place,
				Status:       tmplEvent.Status,
				Link:         tmplEvent.Link,
				Date:         tmplEvent.Date,
				ArtistImgUrl: sql.NullString{String: tmplEvent.ArtistImgUrl, Valid: true},
			},
		},
		{
			"pospond event and mark as resend, when date changes",
			Event{Date: tmplEvent.Date},
			db.Event{
				ID:                 1,
				Date:               tmplEvent.Date.AddDate(0, -1, 0),
				ReportedAtNew:      sql.NullTime{Time: time.Now(), Valid: true},
				ReportedAtUpcoming: sql.NullTime{Time: time.Now(), Valid: true},
			},
			db.Event{
				ID:            1,
				Date:          tmplEvent.Date,
				PostponedDate: sql.NullTime{Time: tmplEvent.Date.AddDate(0, -1, 0), Valid: true},
			},
		},
		{
			"ignore date changes when date < 24 from now",
			Event{Date: time.Now()},
			db.Event{ID: 1, Date: tmplEvent.Date.AddDate(0, -1, 0)},
			db.Event{ID: 1, Date: tmplEvent.Date.AddDate(0, -1, 0)},
		},
		{
			"don't set empty event Name",
			Event{},
			db.Event{ID: 1, Name: "N", Place: "P", Status: "S", Link: "L", ArtistImgUrl: sql.NullString{String: "URL", Valid: true}},
			db.Event{ID: 1, Name: "N", Place: "P", Status: "S", Link: "L", ArtistImgUrl: sql.NullString{String: "URL", Valid: true}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			dbEvent := test.collectedEvent.ToDbEvent(test.dbEvent)

			utils.Debug(dbEvent)
			utils.Debug(test.expEvent)

			assert.Equal(t, test.expEvent, dbEvent)
		})
	}
}
