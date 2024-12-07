package notify

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

const DB_DIALECT = "sqlite3"

type Service struct {
	db     *sql.DB
	Client *whatsmeow.Client
}

func RegisterWhatsApp(ctx context.Context, db *sql.DB, processQRCode func(qrCode string)) error {
	log := waLog.Stdout("Database", LOGLEVEL, true)

	container := sqlstore.NewWithDB(db, DB_DIALECT, log)
	if err := container.Upgrade(); err != nil {
		return err
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return err
	}

	client := whatsmeow.NewClient(deviceStore, log)

	if client.Store.ID != nil {
		return errors.New("An account is already linked, use force to reset.")
	}

	// No ID stored, new login
	qrChan, _ := client.GetQRChannel(ctx)
	err = client.Connect()
	if err != nil {
		return err
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			processQRCode(evt.Code)
			time.Sleep(10 * time.Second)

			return nil
		} else {
			return errors.New("Could not pair: " + evt.Event)
		}
	}

	return errors.New("Could not link device")
}

func ConnectWhatsApp(db *sql.DB, sender string) (*Service, error) {
	log := waLog.Stdout("Database", LOGLEVEL, true)
	container := sqlstore.NewWithDB(db, DB_DIALECT, log)

	jid, err := types.ParseJID(sender)
	if err != nil {
		return nil, err
	}

	deviceStore, _ := container.GetDevice(jid)

	if deviceStore == nil {
		return nil, fmt.Errorf("Could not find device for [%v]", sender)
	}

	client := whatsmeow.NewClient(deviceStore, log)

	if client.Store.ID == nil {
		return nil, fmt.Errorf("No client registered with tel [%v]", sender)
	}

	err = client.Connect()

	if err != nil {
		return nil, err
	}

	return &Service{
		Client: client,
		db:     db,
	}, nil
}

func (s *Service) Send(ctx context.Context, receiver string, message string) error {
	jid, err := types.ParseJID(receiver)
	if err != nil {
		return err
	}

	s.Client.SendMessage(
		ctx,
		jid,
		&waE2E.Message{
			Conversation: proto.String(message),
		},
	)

	return nil
}

func (s *Service) SendImage(ctx context.Context, receiver string, message string, image []byte, mimeType string) error {
	jid, err := types.ParseJID(receiver)
	if err != nil {
		return err
	}

	uploadedImage, err := s.Client.Upload(ctx, image, whatsmeow.MediaImage)
	if err != nil {
		return err
	}

	imageMessage := &waE2E.ImageMessage{
		URL:           &uploadedImage.URL,
		DirectPath:    &uploadedImage.DirectPath,
		MediaKey:      uploadedImage.MediaKey,
		Mimetype:      &mimeType,
		FileEncSHA256: uploadedImage.FileEncSHA256,
		FileSHA256:    uploadedImage.FileSHA256,
		FileLength:    &uploadedImage.FileLength,
		Caption:       &message,
	}

	_, err = s.Client.SendMessage(
		ctx,
		jid,
		&waE2E.Message{
			ImageMessage: imageMessage,
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	return s.Client.GetJoinedGroups()
}

func (s *Service) Close() {
	s.Client.Disconnect()
}
