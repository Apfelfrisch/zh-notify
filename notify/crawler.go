package notify

import (
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

const URL = "https://www.zollhaus-leer.com/veranstaltungen/"

var location = Must(time.LoadLocation("Europe/Berlin"))

type Event struct {
	Name   string
	Place  string
	Date   time.Time
	Status string
	Link   string
}

func CrawlLinks() ([]Event, error) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	var events []Event

	c := colly.NewCollector()

	c.OnScraped(func(r *colly.Response) {
		defer waitGroup.Done()
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
