package collect

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/apfelfrisch/zh-notify/internal/db"

	"github.com/gocolly/colly/v2"
)

type EventSyncCollector interface {
	Init() error
	SetCategory(event *db.Event) error
	SetArtist(event *db.Event) error
	SetArtistUrl(event *db.Event) error
	SetArtistImgUrl(event *db.Event) error
}

type Event struct {
	Name         string
	Place        string
	Date         time.Time
	Status       string
	Link         string
	ArtistImgUrl string
}

func (pe Event) ToDbEvent(dbEvent db.Event) db.Event {
	if dbEvent.ID == 0 {
		dbEvent.Date = pe.Date
		dbEvent.Name = strings.TrimSpace(pe.Name)
		dbEvent.Place = strings.TrimSpace(pe.Place)
		dbEvent.Status = strings.TrimSpace(pe.Status)
		dbEvent.Link = strings.TrimSpace(pe.Link)
		dbEvent.ArtistImgUrl = sql.NullString{String: strings.TrimSpace(pe.ArtistImgUrl), Valid: true}

		return dbEvent
	}

	// If the Event was postpone, reset report it again
	if pe.Date.Sub(time.Now()).Hours() > 24 && math.Abs(pe.Date.Sub(dbEvent.Date).Hours()) > 23.0 {
		dbEvent.ReportedAtUpcoming = sql.NullTime{}
		dbEvent.ReportedAtNew = sql.NullTime{}
		dbEvent.PostponedDate = sql.NullTime{Time: dbEvent.Date, Valid: true}
	}

	if pe.Date.Sub(time.Now()).Hours() > 24 {
		dbEvent.Date = pe.Date
	}

	if strings.TrimSpace(pe.Name) != "" {
		dbEvent.Name = strings.TrimSpace(pe.Name)
	}
	if strings.TrimSpace(pe.Place) != "" {
		dbEvent.Place = strings.TrimSpace(pe.Place)
	}
	if strings.TrimSpace(pe.Status) != "" {
		dbEvent.Status = strings.TrimSpace(pe.Status)
	}
	if strings.TrimSpace(pe.Link) != "" {
		dbEvent.Link = strings.TrimSpace(pe.Link)
	}
	if strings.TrimSpace(pe.ArtistImgUrl) != "" {
		dbEvent.ArtistImgUrl = sql.NullString{String: strings.TrimSpace(pe.ArtistImgUrl), Valid: true}
	}

	return dbEvent
}

func CrawlEvents(url string) ([]Event, error) {
	var waitGroup sync.WaitGroup

	var events []Event

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		waitGroup.Add(1)
	})

	c.OnHTML(".zhcal", func(e *colly.HTMLElement) {
		event := Event{}

		e.ForEachWithBreak(".zhcalband", func(i int, e *colly.HTMLElement) bool {
			event.Name = e.Text
			return false
		})

		e.ForEachWithBreak(".zhcaldate", func(i int, e *colly.HTMLElement) bool {
			event.Date, _ = time.ParseInLocation(
				"20060102",
				e.ChildAttr(".date", "data-showdate"),
				time.Local,
			)
			// Avoid Timezone problems
			event.Date = event.Date.Add(time.Hour * 6)

			return false
		})

		e.ForEachWithBreak(".zhcalvenue", func(i int, e *colly.HTMLElement) bool {
			event.Place = e.Text
			return false
		})

		e.ForEachWithBreak(".zhcalstatus", func(i int, e *colly.HTMLElement) bool {
			event.Status = e.Text
			return false
		})

		e.ForEachWithBreak("a[href]", func(i int, e *colly.HTMLElement) bool {
			event.Link = e.Attr("href")
			if event.Link != "" {
				waitGroup.Add(1)
				go func(link string) {
					defer waitGroup.Done()
					err := e.Request.Visit(link)
					if err != nil {
						fmt.Printf("Failed to visit link %s: %v\n", link, err)
					}
				}(event.Link)
			}
			return false
		})

		events = append(events, event)
	})

	c.OnHTML(".attachment-shows", func(e *colly.HTMLElement) {
		for i := range events {
			if events[i].Link == e.Request.URL.String() {
				events[i].ArtistImgUrl = e.Attr("src")
				break
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		waitGroup.Done()
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}

	waitGroup.Wait()

	return events, nil
}
