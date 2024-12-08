package notify

import (
	"context"
	"database/sql"
	"embed"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
)

const LOGLEVEL = "ERROR"
const DATE_FORMAT = "02.01.06"
const NOTIFY_DAYS_AHEAD = 15

//go:embed "images"
var files embed.FS

var sb strings.Builder

type SendImageParams struct {
	ctx      context.Context
	receiver string
	message  string
	image    []byte
	mimeType string
}

type Driver interface {
	SendWithImage(arg SendImageParams) error
}

type EventRepo interface {
	CreateEvent(ctx context.Context, arg db.CreateEventParams) error
	GetEvent(ctx context.Context, id int64) (db.Event, error)
	GetFreshEvents(ctx context.Context) ([]db.Event, error)
	GetEventsForPeriod(ctx context.Context, arg db.GetEventsForPeriodParams) ([]db.Event, error)
	MarkFreshEventsAsReported(ctx context.Context, arg db.MarkFreshEventsAsReportedParams) error
	MarkUpcomingEventsAsReported(ctx context.Context, arg db.MarkUpcomingEventsAsReportedParams) error
}

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func NewWhatsAppDriver(db *sql.DB, senderJid string) Driver {
	sender, err := ConnectWhatsApp(db, senderJid)

	if err != nil {
		log.Panic(err)
	}

	return sender
}

type notificator struct {
	queries EventRepo
	sender  Driver
}

func NewNotificator(queries *db.Queries, sender Driver) notificator {
	return notificator{queries, sender}
}

func (n notificator) SendMonthlyEvents(ctx context.Context, receiver string) {
	// endOfMonth: end of current month + 15 days
	nm := time.Now().AddDate(0, 0, NOTIFY_DAYS_AHEAD)
	startOfMonth := time.Date(nm.Year(), nm.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	events, _ := n.queries.GetEventsForPeriod(ctx, db.GetEventsForPeriodParams{
		Date:   time.Now(),
		Date_2: endOfMonth,
	})

	for _, event := range events {
		err := n.sender.SendWithImage(SendImageParams{
			ctx:      ctx,
			receiver: receiver,
			message:  buildMessage(event, true),
			image:    getEventImage(event),
			mimeType: "image/jpeg",
		})

		if err != nil {
			continue
		}

		n.queries.MarkUpcomingEventsAsReported(ctx, db.MarkUpcomingEventsAsReportedParams{
			ReportedAtUpcoming: sql.NullTime{Time: time.Now(), Valid: true},
			ID:                 event.ID,
		})
	}
}

func (n notificator) SendFreshEvents(ctx context.Context, receiver string) {
	events, _ := n.queries.GetFreshEvents(ctx)

	for _, event := range events {
		n.sender.SendWithImage(SendImageParams{
			ctx:      ctx,
			receiver: receiver,
			message:  buildMessage(event, false),
			image:    getEventImage(event),
			mimeType: "image/jpeg",
		})

		n.queries.MarkFreshEventsAsReported(ctx, db.MarkFreshEventsAsReportedParams{
			ReportedAtNew: sql.NullTime{Time: time.Now(), Valid: true},
			ID:            event.ID,
		})
	}
}

func buildMessage(event db.Event, withPlace bool) string {
	sb.Reset()
	sb.WriteString(event.Name)
	sb.WriteString("\n\n")
	sb.WriteString(event.Date.Format(DATE_FORMAT))
	// When set, this is the monthly report
	if event.ReportedAtNew.Valid {
		sb.WriteString(" | " + event.Status)
	}
	if withPlace {
		sb.WriteString("\nOrt: ")
		sb.WriteString(event.Place)
	}
	if event.ArtistUrl.Valid {
		sb.WriteString("\nSpotify: " + event.ArtistUrl.String)
	}
	sb.WriteString("\nInfo: ")
	sb.WriteString(event.Link)
	return sb.String()
}

func getEventImage(event db.Event) []byte {
	if !event.ArtistImgUrl.Valid {
		return getFallbackImge(event)
	}

	resp, err := http.Get(event.ArtistImgUrl.String)
	if err != nil {
		return getFallbackImge(event)
	}
	defer resp.Body.Close()

	img, err := io.ReadAll(resp.Body)
	if err != nil {
		return getFallbackImge(event)
	}

	return img
}

func getFallbackImge(event db.Event) []byte {
	switch event.Category.String {
	case "comedy":
		return getFileContent("images/comedy.jpeg")
	case "concert":
		return getFileContent("images/concert.jpeg")
	case "party":
		return getFileContent("images/party.jpeg")
	case "theatre":
		return getFileContent("images/theatre.jpeg")
	case "reading":
		return getFileContent("images/reading.jpeg")
	default:
		return getFileContent("images/fallback.jpeg")
	}
}

func getFileContent(name string) []byte {
	content, err := files.ReadFile(name)

	if err != nil {
		panic("Could not read file [" + name + "] :" + err.Error())
	}

	return content
}
