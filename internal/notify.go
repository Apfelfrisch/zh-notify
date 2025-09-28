package internal

import (
	"bytes"
	"context"
	"database/sql"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/apfelfrisch/zh-notify/assets"
	"github.com/apfelfrisch/zh-notify/internal/db"
	"github.com/apfelfrisch/zh-notify/internal/transport"
	"github.com/apfelfrisch/zh-notify/internal/transport/whatsapp"

	"github.com/disintegration/imaging"
)

const DATE_FORMAT = "02.01.‘06"
const NOTIFY_DAYS_AHEAD = 15
const MAX_IMAGE_SIZE = 500

var sb strings.Builder

func NewNotificator(ctx context.Context, senderJid string) (*Notificator, error) {
	conn, err := sql.Open("sqlite3", "database.sqlite")

	if err != nil {
		return nil, err
	}

	sender, err := whatsapp.Connect(ctx, conn, senderJid)

	if err != nil {
		log.Panic(err)
	}

	return &Notificator{db.NewEventRepoFromConn(conn), sender}, nil
}

type Notificator struct {
	eventRepo db.EventRepository
	sender    transport.Driver
}

func (n Notificator) SendMonthlyEvents(ctx context.Context, receiver string) {
	events, _ := n.eventRepo.GetUpcomingEvents(ctx, time.Now(), NOTIFY_DAYS_AHEAD)

	for _, event := range events {
		mimeType, image := getEventImage(event)

		err := n.sender.SendWithImage(transport.SendImageParams{
			Ctx:      ctx,
			Receiver: receiver,
			Message:  buildMessage(event, true),
			Image:    image,
			MimeType: mimeType,
		})

		if err != nil {
			continue
		}

		event.ReportedAtUpcoming = sql.NullTime{Time: time.Now(), Valid: true}

		n.eventRepo.Save(ctx, event)
	}
}

func (n Notificator) SendFreshEvents(ctx context.Context, receiver string) {
	events, _ := n.eventRepo.GetFreshEvents(ctx)

	for _, event := range events {
		mimeType, image := getEventImage(event)

		n.sender.SendWithImage(transport.SendImageParams{
			Ctx:      ctx,
			Receiver: receiver,
			Message:  buildMessage(event, false),
			Image:    image,
			MimeType: mimeType,
		})

		event.ReportedAtNew = sql.NullTime{Time: time.Now(), Valid: true}

		n.eventRepo.Save(ctx, event)
	}
}

func buildMessage(event db.Event, withStatus bool) string {
	sb.Reset()
	sb.WriteString(event.Name)
	sb.WriteString("\n\n")

	if !event.PostponedDate.Valid {
		sb.WriteString("*")
		sb.WriteString(event.Date.Format(DATE_FORMAT))
		sb.WriteString("*")
	} else {
		sb.WriteString("~")
		sb.WriteString(event.PostponedDate.Time.Format(DATE_FORMAT))
		sb.WriteString("~ : *")
		sb.WriteString(event.Date.Format(DATE_FORMAT))
		sb.WriteString("*")
	}

	if withStatus {
		sb.WriteString(" | ")
		sb.WriteString(event.Status)
	}

	sb.WriteString("\nLocation: ")
	sb.WriteString(event.Place)

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
	content, err := assets.Files.ReadFile(name)

	if err != nil {
		panic("Could not read file [" + name + "] :" + err.Error())
	}

	return content
}
