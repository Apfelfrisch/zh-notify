package cmd

import (
	"context"
	"strings"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/notify"
	"github.com/spf13/cobra"
)

var crawlEventsCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl Events from " + notify.URL,
	Args:  cobra.ExactArgs(0), // Ensure exactly one argument is passed
	RunE: func(cmd *cobra.Command, args []string) error {
		return crawlEvents(cmd.Context(), notify.Must(db.NewSqliteService()).Queries)
	},
}

func crawlEvents(ctx context.Context, queries *db.Queries) error {
	events, err := notify.CrawlLinks()

	if err != nil {
		return err
	}

	for _, event := range events {
		err := queries.CreateEvent(ctx, db.CreateEventParams{
			Name:   strings.TrimSpace(event.Name),
			Place:  strings.TrimSpace(event.Place),
			Status: strings.TrimSpace(event.Status),
			Link:   strings.TrimSpace(event.Link),
			Date:   event.Date,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
