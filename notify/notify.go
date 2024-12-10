package notify

import (
	"context"
	"database/sql"
	"embed"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
)

const LOGLEVEL = "ERROR"
const DATE_FORMAT = "02.01.‘06"
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

func NewWhatsAppDriver(db *sql.DB, senderJid string) Driver {
	sender, err := ConnectWhatsApp(db, senderJid)

	if err != nil {
		log.Panic(err)
	}

	return sender
}

type notificator struct {
	eventRepo EventRepository
	sender    Driver
}

func NewNotificator(eventRepo EventRepository, sender Driver) notificator {
	return notificator{eventRepo, sender}
}

func (n notificator) SendMonthlyEvents(ctx context.Context, receiver string) {
	events, _ := n.eventRepo.GetUpcomingEvents(ctx, time.Now(), NOTIFY_DAYS_AHEAD)

	for _, event := range events {
		imageType := strings.TrimPrefix(filepath.Ext(event.ArtistImgUrl.String), ".")
		if imageType == "" {
			imageType = "jpeg"
		}

		err := n.sender.SendWithImage(SendImageParams{
			ctx:      ctx,
			receiver: receiver,
			message:  buildMessage(event, true),
			image:    getEventImage(event),
			mimeType: "image/" + imageType,
		})

		if err != nil {
			continue
		}

		event.ReportedAtUpcoming = sql.NullTime{Time: time.Now(), Valid: true}

		n.eventRepo.Save(ctx, event)
	}
}

func (n notificator) SendFreshEvents(ctx context.Context, receiver string) {
	events, _ := n.eventRepo.GetFreshEvents(ctx)

	for _, event := range events {
		n.sender.SendWithImage(SendImageParams{
			ctx:      ctx,
			receiver: receiver,
			message:  buildMessage(event, false),
			image:    getEventImage(event),
			mimeType: "image/jpeg",
		})

		event.ReportedAtNew = sql.NullTime{Time: time.Now(), Valid: true}

		n.eventRepo.Save(ctx, event)
	}
}

func buildMessage(event db.Event, withPlace bool) string {
	sb.Reset()
	sb.WriteString(event.Name)
	sb.WriteString("\n\n")

	if !event.PostponedDate.Valid {
		sb.WriteString("*")
		sb.WriteString(event.Date.Format(DATE_FORMAT))
		sb.WriteString("*")
	} else {
		sb.WriteString("~")
		sb.WriteString(event.Date.Format(DATE_FORMAT))
		sb.WriteString("~ : *")
		sb.WriteString(event.Date.Format(DATE_FORMAT))
		sb.WriteString("*")
	}

	// When set, this is the monthly report
	if event.ReportedAtNew.Valid {
		sb.WriteString(" | ")
		sb.WriteString(event.Status)
	}
	if withPlace {
		sb.WriteString("\nOrt: ")
		sb.WriteString(event.Place)
	}
	if event.ArtistUrl.Valid {
		sb.WriteString("\nSpotify: ")
		sb.WriteString(event.ArtistUrl.String)
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
