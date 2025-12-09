package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourusername/flowgres2memos/internal/config"
	"github.com/yourusername/flowgres2memos/internal/migrate"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate journal entries from Flowgress backup to Memos",
	Long: `Extracts all journal entries from the Flowgress SQLite database
and migrates them to your Memos instance, including images and tags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		// Create migrator
		migrator, err := migrate.NewMigrator(cfg, backupPath, dryRun)
		if err != nil {
			return err
		}

		// Run migration
		return migrator.Migrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
