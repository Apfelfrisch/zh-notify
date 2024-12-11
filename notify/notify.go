package notify

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/disintegration/imaging"
)

const LOGLEVEL = "ERROR"
const DATE_FORMAT = "02.01.‘06"
const NOTIFY_DAYS_AHEAD = 15
const MAX_IMAGE_SIZE = 500

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
		mimeType, image := getEventImage(event)

		err := n.sender.SendWithImage(SendImageParams{
			ctx:      ctx,
			receiver: receiver,
			message:  buildMessage(event, true),
			image:    image,
			mimeType: mimeType,
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
		mimeType, image := getEventImage(event)

		n.sender.SendWithImage(SendImageParams{
			ctx:      ctx,
			receiver: receiver,
			message:  buildMessage(event, false),
			image:    image,
			mimeType: mimeType,
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

func getEventImage(event db.Event) (string, []byte) {
	if !event.ArtistImgUrl.Valid {
		return getFallbackImge(event)
	}

	resp, err := http.Get(event.ArtistImgUrl.String)
	if err != nil {
		return getFallbackImge(event)
	}
	defer resp.Body.Close()

	format, err := imaging.FormatFromFilename(event.ArtistImgUrl.String)
	if err != nil {
		format, _ = imaging.FormatFromExtension(".jpeg")
	}

	img, err := imaging.Decode(resp.Body)
	if err != nil {
		return getFallbackImge(event)
	}

	buf := bytes.NewBuffer([]byte{})

	if err := imaging.Encode(buf, imaging.Fit(img, MAX_IMAGE_SIZE, MAX_IMAGE_SIZE, imaging.Lanczos), format); err != nil {
		return getFallbackImge(event)
	}

	return mime.TypeByExtension("." + format.String()), buf.Bytes()
}

func getFallbackImge(event db.Event) (string, []byte) {
	switch event.Category.String {
	case "comedy":
		return "image/jpeg", getFileContent("images/comedy.jpeg")
	case "concert":
		return "image/jpeg", getFileContent("images/concert.jpeg")
	case "party":
		return "image/jpeg", getFileContent("images/party.jpeg")
	case "theatre":
		return "image/jpeg", getFileContent("images/theatre.jpeg")
	case "reading":
		return "image/jpeg", getFileContent("images/reading.jpeg")
	default:
		return "image/jpeg", getFileContent("images/fallback.jpeg")
	}
}

func getFileContent(name string) []byte {
	content, err := files.ReadFile(name)

	if err != nil {
		panic("Could not read file [" + name + "] :" + err.Error())
	}

	return content
}
