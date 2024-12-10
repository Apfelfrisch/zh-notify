package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/notify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var notifyCmd = &cobra.Command{
	Use:   "notify [upcoming|fresh]",
	Short: "Broadcast Zollhaus Events",
	Args:  validateNotifyArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		service, err := db.NewSqliteService()
		if err != nil {
			return err
		}

		switch args[0] {
		case "fresh":
			return notifyFresh(cmd.Context(), service)
		case "upcoming":
			return notifyMonthly(cmd.Context(), service)
		}

		return errors.New("unexpected error occurred")
	},
}

func validateNotifyArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("exactly one argument is required: 'upcoming' or 'fresh'")
	}
	if args[0] != "upcoming" && args[0] != "fresh" {
		return fmt.Errorf("invalid argument: %s. Allowed values are 'upcoming' or 'fresh'", args[0])
	}
	return nil
}

func notifyMonthly(ctx context.Context, service *db.Service) error {
	senderJid := viper.GetString("SENDER_JID")
	if senderJid == "" {
		return errors.New("Could not read SENDER_JID from env")
	}

	monthlyChannel := viper.GetString("MONTHLY_CHANNEL_JID")
	if monthlyChannel == "" {
		return errors.New("Could not read SENDER_JID from env")
	}

	notificator := notify.NewNotificator(
		notify.NewDbEventRepo(service.Queries),
		notify.NewWhatsAppDriver(service.Db, senderJid),
	)

	notificator.SendMonthlyEvents(ctx, monthlyChannel)

	return nil
}

func notifyFresh(ctx context.Context, service *db.Service) error {
	senderJid := viper.GetString("SENDER_JID")
	if senderJid == "" {
		return errors.New("Could not read SENDER_JID from env")
	}

	monthlyChannel := viper.GetString("MONTHLY_CHANNEL_JID")
	if monthlyChannel == "" {
		return errors.New("Could not read MONTHLY_CHANNEL_JID from env")
	}

	justAddedChannel := viper.GetString("NEW_EVENTS_CHANNEL_JID")
	if monthlyChannel == "" {
		return errors.New("Could not read NEW_EVENTS_CHANNEL_JID from env")
	}

	notificator := notify.NewNotificator(
		notify.NewDbEventRepo(service.Queries),
		notify.NewWhatsAppDriver(service.Db, senderJid),
	)

	notificator.SendFreshEvents(ctx, justAddedChannel)

	return nil
}
