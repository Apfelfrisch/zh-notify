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

func (s *Service) SendWithImage(arg SendImageParams) error {
	jid, err := types.ParseJID(arg.receiver)

	if err != nil {
		return err
	}

	imageMessage, err := func() (*waE2E.ImageMessage, error) {
		imageArg := imageParams{
			ctx:      arg.ctx,
			message:  arg.message,
			image:    arg.image,
			mimeType: arg.mimeType,
		}
		if jid.Server == "newsletter" {
			return s.buildChannelImageMessage(imageArg)
		}
		return s.buildImageMessage(imageArg)
	}()
	if err != nil {
		return err
	}

	_, err = s.Client.SendMessage(
		arg.ctx,
		jid,
		&waE2E.Message{
			ImageMessage: imageMessage,
		},
		// If channel pictures deos not work as expected try this
		// whatsmeow.SendRequestExtra{
		// 	MediaHandle: uploadedImage.Handle,
		// }
	)

	if err != nil {
		return err
	}

	return nil
}

type imageParams struct {
	ctx      context.Context
	message  string
	image    []byte
	mimeType string
}

func (s *Service) buildImageMessage(arg imageParams) (*waE2E.ImageMessage, error) {
	uploadedImage, err := s.Client.Upload(arg.ctx, arg.image, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	imageMessage := &waE2E.ImageMessage{
		URL:           &uploadedImage.URL,
		DirectPath:    &uploadedImage.DirectPath,
		MediaKey:      uploadedImage.MediaKey,
		Mimetype:      &arg.mimeType,
		FileEncSHA256: uploadedImage.FileEncSHA256,
		FileSHA256:    uploadedImage.FileSHA256,
		FileLength:    &uploadedImage.FileLength,
		Caption:       &arg.message,
	}

	return imageMessage, nil
}

func (s *Service) buildChannelImageMessage(arg imageParams) (*waE2E.ImageMessage, error) {
	uploadedImage, err := s.Client.UploadNewsletter(arg.ctx, arg.image, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	imageMessage := &waE2E.ImageMessage{
		Caption:  &arg.message,
		Mimetype: &arg.mimeType,

		URL:        &uploadedImage.URL,
		DirectPath: &uploadedImage.DirectPath,
		FileSHA256: uploadedImage.FileSHA256,
		FileLength: &uploadedImage.FileLength,
	}

	return imageMessage, nil
}

func (s *Service) GetGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	return s.Client.GetJoinedGroups()
}

func (s *Service) Close() {
	s.Client.Disconnect()
}
