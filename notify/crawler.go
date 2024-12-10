package notify

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/util"
	"github.com/gocolly/colly/v2"
)

const URL = "https://www.zollhaus-leer.com/veranstaltungen/"

var location = util.Must(time.LoadLocation("Europe/Berlin"))

type CrawledEvent struct {
	Name         string
	Place        string
	Date         time.Time
	Status       string
	Link         string
	ArtistImgUrl string
}

func (e CrawledEvent) ToDbEvent(dbEvent db.Event) db.Event {
	if dbEvent.ID == 0 {
		dbEvent.Date = e.Date
		dbEvent.Name = strings.TrimSpace(e.Name)
		dbEvent.Place = strings.TrimSpace(e.Place)
		dbEvent.Status = strings.TrimSpace(e.Status)
		dbEvent.Link = strings.TrimSpace(e.Link)
		dbEvent.ArtistImgUrl = sql.NullString{String: strings.TrimSpace(e.ArtistImgUrl), Valid: true}

		return dbEvent
	}

	// If the Event was postpone, reset repoted it again
	if e.Date.Sub(time.Now()).Hours() > 24 && math.Abs(e.Date.Sub(dbEvent.Date).Hours()) > 23.0 {
		dbEvent.ReportedAtUpcoming = sql.NullTime{}
		dbEvent.ReportedAtNew = sql.NullTime{}
		dbEvent.PostponedDate = sql.NullTime{Time: dbEvent.Date, Valid: true}
	}

	if e.Date.Sub(time.Now()).Hours() > 24 {
		dbEvent.Date = e.Date
	}
	if strings.TrimSpace(e.Place) != "" {
		dbEvent.Name = strings.TrimSpace(e.Name)
	}
	if strings.TrimSpace(e.Place) != "" {
		dbEvent.Place = strings.TrimSpace(e.Place)
	}
	if strings.TrimSpace(e.Status) != "" {
		dbEvent.Status = strings.TrimSpace(e.Status)
	}
	if strings.TrimSpace(e.Link) != "" {
		dbEvent.Link = strings.TrimSpace(e.Link)
	}
	if strings.TrimSpace(e.ArtistImgUrl) != "" {
		dbEvent.ArtistImgUrl = sql.NullString{String: strings.TrimSpace(e.ArtistImgUrl), Valid: true}
	}

	return dbEvent
}

func CrawlLinks() ([]CrawledEvent, error) {
	var waitGroup sync.WaitGroup

	var events []CrawledEvent

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		waitGroup.Add(1)
	})

	c.OnHTML(".zhcal", func(e *colly.HTMLElement) {
		event := CrawledEvent{}

		e.ForEachWithBreak(".zhcalband", func(i int, e *colly.HTMLElement) bool {
			event.Name = e.Text
			return false
		})

		e.ForEachWithBreak(".zhcaldate", func(i int, e *colly.HTMLElement) bool {
			event.Date, _ = time.ParseInLocation(
				"20060102",
				e.ChildAttr(".date", "data-showdate"),
				location,
			)

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

	if err := c.Visit(URL); err != nil {
		return nil, err
	}

	waitGroup.Wait()

	return events, nil
}

func TestCrawlLinks(t *testing.T) {
	t.Run("test crawl zollhaus-leer", func(t *testing.T) {
		t.Errorf("Done")
	})
}
