package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourusername/flowgres2memos/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize flowgres2memos configuration",
	Long: `Creates the configuration directory at ~/.flowgres2memos/ and
generates a sample config.yaml file for you to edit with your credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir, err := config.GetConfigDir()
		if err != nil {
			return err
		}

		// Create config directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		configPath := filepath.Join(configDir, "config.yaml")

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("⚠️  Config file already exists at: %s\n", configPath)
			fmt.Println("Edit this file to update your configuration.")
			return nil
		}

		// Create sample config file
		sampleConfig := `# Flowgres to Memos Configuration

# Memos Instance URL
# Your self-hosted Memos instance URL (without trailing slash)
# Example: https://memos.example.com or http://localhost:5230
memos_url: "https://memos.example.com"

# Memos Access Token
# Generate a token in Memos: Settings -> Access Tokens
# The token should have permission to create memos
memos_token: "YOUR_MEMOS_ACCESS_TOKEN_HERE"
`

		if err := os.WriteFile(configPath, []byte(sampleConfig), 0600); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Println("✅ Configuration initialized successfully!")
		fmt.Printf("📁 Config directory: %s\n", configDir)
		fmt.Printf("📄 Config file: %s\n", configPath)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Edit the config file with your Memos server URL and access token")
		fmt.Println("2. Run: flowgres2memos migrate --dry-run  (to test)")
		fmt.Println("3. Run: flowgres2memos migrate  (to perform the actual migration)")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
