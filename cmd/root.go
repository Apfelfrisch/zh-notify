package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zh-notify",
	Short: "Crawl zollhaus-leer.de for new event, and pulish them via whatspp.",
	Long:  "Crawl zollhaus-leer.de for new event, and pulish them via whatspp.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	rootCmd.AddCommand(crawlEventsCmd)
	rootCmd.AddCommand(linkAccountCmd)
	rootCmd.AddCommand(notifyCmd)
	rootCmd.AddCommand(updateMetadataCmd)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
