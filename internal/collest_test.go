package internal

import (
	"database/sql"
	"testing"

	"github.com/apfelfrisch/zh-notify/internal/db"
	"github.com/apfelfrisch/zh-notify/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestCollectEventData(t *testing.T) {
	var tests = []struct {
		name           string
		givenEvent     db.Event
		collectedEvent db.Event
		expectedEvent  db.Event
	}{
		{
			"test set all metadata",
			db.Event{
				Artist:       sql.NullString{},
				Category:     sql.NullString{},
				ArtistUrl:    sql.NullString{},
				ArtistImgUrl: sql.NullString{},
			},
			db.Event{
				Category:     sql.NullString{String: "collected catergory", Valid: true},
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
			db.Event{
				Category:     sql.NullString{String: "collected catergory", Valid: true},
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
		},
		{
			"test set no metadata",
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
			db.Event{
				Category:     sql.NullString{String: "collected catergory", Valid: true},
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
		},
		{
			"test set only the artist",
			db.Event{
				Artist:       sql.NullString{},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				Category:     sql.NullString{String: "Collected Catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
		},
		{
			"test set only the catergory",
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{},
				ArtistUrl:    sql.NullString{String: "artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				Category:     sql.NullString{String: "Collected Catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{String: "Collected Catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
		},
		{
			"test set only the artist-url",
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				Category:     sql.NullString{String: "Collected Catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "artist-img-url", Valid: true},
			},
		},
		{
			"test set only the artist-img-url",
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{},
			},
			db.Event{
				Artist:       sql.NullString{String: "collected.artist", Valid: true},
				Category:     sql.NullString{String: "Collected Catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
			db.Event{
				Artist:       sql.NullString{String: "artist", Valid: true},
				Category:     sql.NullString{String: "catergory", Valid: true},
				ArtistUrl:    sql.NullString{String: "collected.artist.url", Valid: true},
				ArtistImgUrl: sql.NullString{String: "collected.artist-img-url", Valid: true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sc := SyncCollector{
				service: InMemoryEventSyncCollector{tmplEvent: test.collectedEvent},
			}

			utils.Debug(test.givenEvent)

			sc.Sync(&test.givenEvent)

			utils.Debug(test.givenEvent)

			assert.Equal(t, test.expectedEvent, test.givenEvent)
		})
	}
}

type InMemoryEventSyncCollector struct {
	tmplEvent db.Event
}

func (ic InMemoryEventSyncCollector) Init() error {
	return nil
}

func (ic InMemoryEventSyncCollector) SetCategory(event *db.Event) error {
	event.Category = ic.tmplEvent.Category

	return nil
}

func (ic InMemoryEventSyncCollector) SetArtist(event *db.Event) error {
	event.Artist = ic.tmplEvent.Artist

	return nil
}

func (ic InMemoryEventSyncCollector) SetArtistUrl(event *db.Event) error {
	event.ArtistUrl = ic.tmplEvent.ArtistUrl

	return nil
}

func (ic InMemoryEventSyncCollector) SetArtistImgUrl(event *db.Event) error {
	event.ArtistImgUrl = ic.tmplEvent.ArtistImgUrl

	return nil
}
