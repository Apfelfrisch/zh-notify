package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/notify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

var updateMetadataCmd = &cobra.Command{
	Use:   "meta",
	Short: "Get Metadata for new Events",
	Args:  cobra.ExactArgs(0), // Ensure exactly one argument is passed
	RunE: func(cmd *cobra.Command, args []string) error {
		chatGptToken := viper.GetString("CHATGPT_TOKEN")
		if chatGptToken == "" {
			return errors.New("Could not read SENDER_JID from env")
		}

		spotifyId := viper.GetString("SPOTIFY_ID")
		if chatGptToken == "" {
			return errors.New("Could not read SENDER_JID from env")
		}

		sporitySecret := viper.GetString("SPOTIFY_SECRET")
		if chatGptToken == "" {
			return errors.New("Could not read SENDER_JID from env")
		}

		service, err := db.NewSqliteService()
		if err != nil {
			return err
		}

		return updateMetadata(
			cmd.Context(),
			notify.NewDbEventRepo(service.Queries),
			notify.NewMetaDataService(chatGptToken, clientcredentials.Config{
				ClientID:     spotifyId,
				ClientSecret: sporitySecret,
				TokenURL:     spotify.TokenURL,
			}),
		)
	},
}

func updateMetadata(ctx context.Context, eventRepo notify.EventRepository, service notify.MetaDataService) error {
	events, err := eventRepo.GetNakedEvents(ctx)

	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	if err := service.Init(); err != nil {
		return err
	}

	for _, event := range events {
		if err := service.SetCategory(&event); err != nil {
			return err
		}

		if !event.Artist.Valid {
			if err := service.SetArtist(&event); err != nil {
				return err
			}
		}

		if err := service.SetArtistUrl(&event); err != nil {
			fmt.Println(err)
		}

		if err := service.SetArtistImgUrl(&event); err != nil {
			fmt.Println(err)
		}

		eventRepo.Save(ctx, event)
	}

	return nil
}
