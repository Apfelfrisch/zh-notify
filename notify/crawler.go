package notify

import (
	"database/sql"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/util"
	"github.com/gocolly/colly/v2"
)

const URL = "https://www.zollhaus-leer.com/veranstaltungen/"

var location = util.Must(time.LoadLocation("Europe/Berlin"))

type CrawledEvent struct {
	Name   string
	Place  string
	Date   time.Time
	Status string
	Link   string
}

func (e CrawledEvent) ToDbEvent(dbEvent db.Event) db.Event {

	// If the Event was postpone, reset repoted it again
	if math.Abs(e.Date.Sub(dbEvent.Date).Hours()) > 23.0 {
		dbEvent.ReportedAtUpcoming = sql.NullTime{}
		dbEvent.ReportedAtNew = sql.NullTime{}
		dbEvent.PostponedDate = sql.NullTime{Time: dbEvent.Date, Valid: true}
	}

	dbEvent.Date = e.Date
	dbEvent.Name = strings.TrimSpace(e.Name)
	dbEvent.Place = strings.TrimSpace(e.Place)
	dbEvent.Status = strings.TrimSpace(e.Status)
	dbEvent.Link = strings.TrimSpace(e.Link)

	return dbEvent
}

func CrawlLinks() ([]CrawledEvent, error) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	var events []CrawledEvent

	c := colly.NewCollector()

	c.OnScraped(func(r *colly.Response) {
		defer waitGroup.Done()
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
			return false
		})

		events = append(events, event)
	})

	if err := c.Visit(URL); err != nil {
		return nil, err
	}

	waitGroup.Wait()

	return events, nil
}
