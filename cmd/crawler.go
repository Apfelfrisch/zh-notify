package cmd

import (
	"context"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/notify"
	"github.com/apfelfrisch/zh-notify/util"
	"github.com/spf13/cobra"
)

var crawlEventsCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl Events from " + notify.URL,
	Args:  cobra.ExactArgs(0), // Ensure exactly one argument is passed
	RunE: func(cmd *cobra.Command, args []string) error {
		events, err := notify.CrawlLinks()

		if err != nil {
			return err
		}

		return saveEvents(cmd.Context(), notify.NewDbEventRepo(util.Must(db.NewSqliteService()).Queries), events)
	},
}

func saveEvents(ctx context.Context, eventRepo notify.EventRepository, events []notify.CrawledEvent) error {
	for _, crawledEvent := range events {
		dbEvent, _ := eventRepo.GetByLink(ctx, crawledEvent.Link)

		if err := eventRepo.Save(ctx, crawledEvent.ToDbEvent(dbEvent)); err != nil {
			return err
		}
	}

	return nil
}
