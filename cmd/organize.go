package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thiagozs/go-download-organizer/internal/organizer"
)

var (
	source string
	dryRun bool
)

var organizeCmd = &cobra.Command{
	Use:   "organize",
	Short: "Organize files in a directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := organizer.Options{
			Source: source,
			DryRun: dryRun,
		}

		return organizer.Run(opts)
	},
}

func init() {
	rootCmd.AddCommand(organizeCmd)

	organizeCmd.Flags().StringVarP(&source, "source", "s", "./Downloads", "Source directory")
	organizeCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without moving files")
}
