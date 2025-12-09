# Flowgress to Memos Migration Tool

Migrate your journal entries from a Flowgress backup to a Memos server.

## Features

- ✅ Extracts all entries from Flowgress SQLite database
- ✅ Preserves original timestamps
- ✅ Uploads images to Memos as resources
- ✅ Converts topics and categories to tags
- ✅ Supports dry-run mode for testing
- ✅ Progress tracking and detailed statistics
- ✅ Written in Go for fast, reliable performance

## Prerequisites

- Go 1.21 or higher (for building from source)
- A running Memos server with API access
- A Flowgress backup folder

## Installation

### Option 1: Build from Source

```bash
# Install dependencies and build
go mod download
go build -o flowgres2memos
```

### Option 2: Download Binary

Download the pre-built binary from the releases page (if available).

## Setup

1. **Create configuration file:**
```bash
cp config.example.yaml config.yaml
```

2. **Edit `config.yaml` with your Memos server details:**
```yaml
# Memos Instance URL
memos_url: "https://memos.example.com"

# Memos Access Token
memos_token: "YOUR_MEMOS_ACCESS_TOKEN_HERE"
```

### Getting Your Memos Access Token

1. Log in to your Memos server
2. Go to Settings → Access Tokens
3. Create a new token
4. Copy the token to your `config.yaml`

## Usage

### Basic Migration

```bash
./flowgres2memos migrate
```

### Dry Run (Recommended First)

Test the migration without uploading to Memos:

```bash
./flowgres2memos migrate --dry-run
```

This saves all memos to `./dry-run-output/` for review.

### Custom Backup Path

If your backup is in a different location:

```bash
./flowgres2memos migrate --backup /path/to/backup
```

### Custom Config File

```bash
./flowgres2memos migrate --config /path/to/config.yaml
```

### All Options

```bash
./flowgres2memos migrate --help
```

## Commands

- `migrate` - Migrate entries from Flowgress to Memos
- `version` - Show version information
- `help` - Show help for any command

## Environment Variables

You can also set configuration via environment variables:

```bash
export FLOWGRES2MEMOS_MEMOS_URL="https://memos.example.com"
export FLOWGRES2MEMOS_MEMOS_TOKEN="your-token-here"
./flowgres2memos migrate
```

## Features in Detail

### Image Upload

- Images are automatically uploaded to Memos as resources
- Resource URLs are embedded in the memo content
- Failed uploads are logged but don't stop the migration

### Tags

- All memos receive the `#flowgres` tag
- Categories become tags (e.g., `#Health`, `#Work`)
- Spaces in names are replaced with hyphens
- The "Lifegoal" category is excluded

### Timestamp Preservation

- Original creation dates from Flowgress are preserved
- Uses Memos API to set the `displayTime` field
- Entries appear in Memos with their original dates

## Troubleshooting

### "Failed to open database"
- Check that the backup path is correct
- Ensure `databases/local_cordova_db.sqlite` exists in the backup folder

### "memos_url is required"
- Make sure `config.yaml` exists with proper settings
- Or set environment variables

### Image upload failures
- Check that image files exist in the backup
- Verify Memos server has sufficient storage
- The migration continues even if some images fail

## Development

### Project Structure

```
flowgres2memos/
├── cmd/               # CLI commands
│   ├── root.go       # Root command setup
│   ├── migrate.go    # Migration command
│   └── version.go    # Version command
├── internal/
│   ├── config/       # Configuration loading
│   ├── flowgres/     # Flowgres database client
│   ├── memos/        # Memos API client
│   └── migrate/      # Migration orchestration
├── main.go           # Entry point
├── go.mod            # Dependencies
└── config.example.yaml
```

### Running Tests

```bash
go test ./...
```

## License

MIT License - See LICENSE file for details
