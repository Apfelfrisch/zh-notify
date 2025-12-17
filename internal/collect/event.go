package collect

import (
	"database/sql"
	"fmt"
	"html"
	"math"
	"regexp"
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
	if time.Until(pe.Date).Hours() > 24 && math.Abs(pe.Date.Sub(dbEvent.Date).Hours()) > 23.0 {
		dbEvent.ReportedAtUpcoming = sql.NullTime{}
		dbEvent.ReportedAtNew = sql.NullTime{}
		dbEvent.PostponedDate = sql.NullTime{Time: dbEvent.Date, Valid: true}
	}

	if time.Until(pe.Date).Hours() > 24 {
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

	c.OnHTML(".elementor-6082", func(e *colly.HTMLElement) {
		event := Event{}

		e.ForEachWithBreak("h3.elementor-heading-title", func(i int, e *colly.HTMLElement) bool {
			event.Name = e.Text
			return false
		})

		// Find the date string in the heading
		e.ForEachWithBreak(".elementor-heading-title", func(i int, el *colly.HTMLElement) bool {
			text := strings.TrimSpace(el.Text)

			re := regexp.MustCompile(`(?m)^(Mo|Di|Mi|Do|Fr|Sa|So)\., \d{2}\.\d{2}\.\d{4}$`)
			matches := re.FindAllString(text, -1)

			for _, match := range matches {
				parsedDate, err := time.ParseInLocation("02.01.2006", match[5:], time.Local)

				if err == nil {
					event.Date = parsedDate.Add(time.Hour * 6)
					return false
				}

			}
			return true
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

	c.OnHTML("body.single-event", func(e *colly.HTMLElement) {
		for i := range events {
			if events[i].Link != e.Request.URL.String() {
				continue
			}

			e.ForEachWithBreak("img.attachment-large", func(_ int, el *colly.HTMLElement) bool {
				events[i].ArtistImgUrl = el.Attr("data-lazy-src")
				if events[i].ArtistImgUrl == "" {
					events[i].ArtistImgUrl = el.Attr("src")
				}
				return false
			})

			e.ForEachWithBreak("li.elementor-icon-list-item", func(_ int, li *colly.HTMLElement) bool {
				hasMapPin := false
				li.ForEach("span.elementor-icon-list-icon i", func(_ int, icon *colly.HTMLElement) {
					if icon.Attr("class") == "fad fa-map-pin" {
						hasMapPin = true
					}
				})
				if hasMapPin {
					events[i].Place = li.ChildText("span.elementor-icon-list-text")
					return false
				}
				return true
			})

			e.ForEachWithBreak("div.elementor-element-5bb6689 .elementor-button-text", func(_ int, btn *colly.HTMLElement) bool {
				events[i].Status = html.UnescapeString(btn.Text)
				return false
			})
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
