package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/apfelfrisch/zh-notify/notify"
	"github.com/mdp/qrterminal"
	"github.com/spf13/cobra"
)

var linkAccountCmd = &cobra.Command{
	Use:   "link",
	Short: "Link a whatsapp account, to send events",
	Args:  cobra.ExactArgs(0), // Ensure exactly one argument is passed
	RunE: func(cmd *cobra.Command, args []string) error {
		return linkAccount(cmd.Context(), notify.Must(db.NewSqliteService()).Db)
	},
}

func linkAccount(ctx context.Context, db *sql.DB) error {
	err := notify.RegisterWhatsApp(ctx, db, func(qrCode string) {
		qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout)
	})

	if err != nil {
		return err
	}

	fmt.Println("Account linked successful")

	return nil
}
