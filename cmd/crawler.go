package cmd

import (
	"context"

	"github.com/apfelfrisch/zh-notify/internal"
	"github.com/apfelfrisch/zh-notify/internal/collect"
	"github.com/apfelfrisch/zh-notify/internal/db"
	"github.com/spf13/cobra"
)

var crawlEventsCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl Events from " + internal.URL,
	Args:  cobra.ExactArgs(0), // Ensure exactly one argument is passed
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := db.NewDbEventRepo()

		if err != nil {
			return err
		}

		events, err := internal.CollectNewEvents()

		if err != nil {
			return err
		}

		return saveEvents(cmd.Context(), repo, events)
	},
}

func saveEvents(ctx context.Context, eventRepo db.EventRepository, events []collect.Event) error {
	for _, crawledEvent := range events {
		dbEvent, _ := eventRepo.GetByLink(ctx, crawledEvent.Link)

		if err := eventRepo.Save(ctx, crawledEvent.ToDbEvent(dbEvent)); err != nil {
			return err
		}
	}

	return nil
}
