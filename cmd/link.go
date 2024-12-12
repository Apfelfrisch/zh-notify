package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/apfelfrisch/zh-notify/internal/db"
	"github.com/apfelfrisch/zh-notify/internal/transport/whatsapp"
	"github.com/mdp/qrterminal"
	"github.com/spf13/cobra"
)

var linkAccountCmd = &cobra.Command{
	Use:   "link",
	Short: "Link a whatsapp account, to send events",
	Args:  cobra.ExactArgs(0), // Ensure exactly one argument is passed
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := db.NewSqliteConn()

		if err != nil {
			return err
		}

		return linkAccount(cmd.Context(), conn)
	},
}

func linkAccount(ctx context.Context, db *sql.DB) error {
	err := whatsapp.Register(ctx, db, func(qrCode string) {
		qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout)
	})

	if err != nil {
		return err
	}

	fmt.Println("Account linked successful")

	return nil
}
