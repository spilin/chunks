package main

import (
	"chunks/cmd"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chunks",
		Short: "Analize chunks",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.AddCommand(cmd.ShowChunkAuthorsCmd)
	rootCmd.AddCommand(cmd.FeedChunkAuthorsCmd)
	rootCmd.AddCommand(cmd.CollectChunkAuthorsCmd)
	rootCmd.AddCommand(cmd.FeedChunksAvailabilityCmd)

	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
