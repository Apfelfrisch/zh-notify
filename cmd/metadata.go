package cmd

import (
	"context"
	"errors"

	"github.com/apfelfrisch/zh-notify/internal"
	"github.com/apfelfrisch/zh-notify/internal/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		repo, err := db.NewDbEventRepo()
		if err != nil {
			return err
		}

		return updateMetadata(
			cmd.Context(),
			repo,
			internal.NewSyncEventCollector(chatGptToken, spotifyId, sporitySecret),
		)
	},
}

func updateMetadata(ctx context.Context, eventRepo db.EventRepository, service *internal.SyncCollector) error {
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
		if err := service.Sync(&event); err != nil {
			return err
		}

		eventRepo.Save(ctx, event)
	}

	return nil
}
