package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	dryRun     bool
	backupPath string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "flowgres2memos",
	Short: "Migrate journal entries from Flowgress to Memos",
	Long: `flowgres2memos is a CLI tool to migrate your journal entries from a Flowgress backup to Memos.
It extracts entries from the SQLite database, uploads images, and preserves timestamps.`,
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flowgres2memos/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "run without actually creating memos (saves to ./dry-run-output/)")
	rootCmd.PersistentFlags().StringVar(&backupPath, "backup", "./flowgress-backup_251209_202529", "path to Flowgress backup directory")
}
