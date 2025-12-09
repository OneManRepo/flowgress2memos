package migrate

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/yourusername/flowgres2memos/internal/config"
	"github.com/yourusername/flowgres2memos/internal/flowgres"
	"github.com/yourusername/flowgres2memos/internal/memos"
)

// Migrator coordinates the migration from Flowgres to Memos
type Migrator struct {
	flowgresClient *flowgres.Client
	memosClient    *memos.Client
	config         *config.Config
	backupPath     string
	dryRun         bool
}

// NewMigrator creates a new Migrator
func NewMigrator(cfg *config.Config, backupPath string, dryRun bool) (*Migrator, error) {
	flowgresClient, err := flowgres.NewClient(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Flowgres client: %w", err)
	}

	memosClient := memos.NewClient(cfg.MemosURL, cfg.MemosToken)

	return &Migrator{
		flowgresClient: flowgresClient,
		memosClient:    memosClient,
		config:         cfg,
		backupPath:     backupPath,
		dryRun:         dryRun,
	}, nil
}

// Migrate performs the migration from Flowgres to Memos
func (m *Migrator) Migrate() error {
	defer m.flowgresClient.Close()

	log.Println("=" + strings.Repeat("=", 79))
	log.Println("🚀 FLOWGRES TO MEMOS MIGRATION")
	log.Println("=" + strings.Repeat("=", 79))
	log.Println()

	if m.dryRun {
		log.Println("⚠️  DRY RUN MODE - No data will be uploaded")
		log.Println()
	}

	// Extract data
	log.Println("📂 Extracting data from Flowgres backup...")

	categories, err := m.flowgresClient.GetCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	topics, err := m.flowgresClient.GetTopics()
	if err != nil {
		return fmt.Errorf("failed to get topics: %w", err)
	}

	entries, err := m.flowgresClient.GetAllEntries()
	if err != nil {
		return fmt.Errorf("failed to get entries: %w", err)
	}

	log.Printf("✅ Found %d categories\n", len(categories))
	log.Printf("✅ Found %d topics\n", len(topics))
	log.Printf("✅ Found %d entries\n", len(entries))
	log.Println()

	// Display categories
	log.Println("📋 Categories:")
	for _, name := range categories {
		log.Printf("   • %s\n", name)
	}
	log.Println()

	// Group entries by date
	entriesByDate := m.groupEntriesByDate(entries)

	log.Printf("🔄 Migrating %d entries grouped into %d days...\n", len(entries), len(entriesByDate))
	log.Println()

	successCount := 0
	failedCount := 0
	uploadedImages := 0
	failedImages := 0

	// Create progress bar
	bar := progressbar.Default(int64(len(entriesByDate)), "Migrating days")

	for date, dayEntries := range entriesByDate {
		if err := m.migrateDayEntries(date, dayEntries, topics, &uploadedImages, &failedImages); err != nil {
			log.Printf("❌ Failed to migrate entries for %s: %v\n", date, err)
			failedCount += len(dayEntries)
		} else {
			successCount += len(dayEntries)
		}
		bar.Add(1)

		// Rate limiting
		if !m.dryRun {
			time.Sleep(500 * time.Millisecond)
		}
	}

	bar.Finish()

	// Print statistics
	log.Println()
	log.Println("=" + strings.Repeat("=", 79))
	log.Println("📊 MIGRATION COMPLETE")
	log.Println("=" + strings.Repeat("=", 79))
	log.Printf("Total entries:     %d\n", len(entries))
	log.Printf("Grouped into days: %d\n", len(entriesByDate))
	log.Printf("Migrated:          %d ✅\n", successCount)
	log.Printf("Failed:            %d ❌\n", failedCount)
	log.Printf("Images uploaded:   %d\n", uploadedImages)
	log.Printf("Images failed:     %d\n", failedImages)
	log.Println("=" + strings.Repeat("=", 79))

	if m.dryRun {
		log.Println()
		log.Println("Check ./dry-run-output/ for the generated markdown files")
	}

	return nil
}

// groupEntriesByDate groups entries by their date (YYYY-MM-DD)
func (m *Migrator) groupEntriesByDate(entries []*flowgres.Entry) map[string][]*flowgres.Entry {
	grouped := make(map[string][]*flowgres.Entry)

	for _, entry := range entries {
		createdTime, err := parseTimestamp(entry.CreatedAt)
		if err != nil {
			log.Printf("⚠️  Warning: Could not parse timestamp %s: %v\n", entry.CreatedAt, err)
			continue
		}

		dateKey := createdTime.Format("2006-01-02")
		grouped[dateKey] = append(grouped[dateKey], entry)
	}

	return grouped
}

// migrateDayEntries migrates all entries for a single day as one memo
func (m *Migrator) migrateDayEntries(date string, entries []*flowgres.Entry, topics map[string]*flowgres.Topic,
	uploadedImages, failedImages *int) error {

	if len(entries) == 0 {
		return nil
	}

	// Parse the date to get the timestamp (use first entry's time for the memo timestamp)
	firstEntry := entries[0]
	createdTime, err := parseTimestamp(firstEntry.CreatedAt)
	if err != nil {
		log.Printf("⚠️  Warning: Could not parse timestamp %s: %v\n", firstEntry.CreatedAt, err)
		createdTime = time.Now()
	}

	// Build combined content for all entries of the day
	var contentParts []string
	allTags := make(map[string]bool)
	topicNames := make(map[string]string) // Map to track first topic name

	// Add flowgress tag to all memos
	allTags["#flowgress"] = true

	// Collect all topics and find the most common one (or first one)
	for _, entry := range entries {
		topic := topics[entry.TopicID]
		if topic != nil && topic.Name != "" {
			if _, exists := topicNames[topic.Name]; !exists {
				topicNames[topic.Name] = topic.Name
			}
		}
	}

	// Add the first topic name as main header at the beginning
	if len(topicNames) > 0 {
		// Get first topic name
		for _, name := range topicNames {
			contentParts = append(contentParts, fmt.Sprintf("# %s", name))
			contentParts = append(contentParts, "")
			break // Only add the first one
		}
	}

	for _, entry := range entries {
		entryTime, err := parseTimestamp(entry.CreatedAt)
		if err != nil {
			log.Printf("⚠️  Warning: Could not parse timestamp %s: %v\n", entry.CreatedAt, err)
			continue
		}

		// Get topic information
		topic := topics[entry.TopicID]
		if topic == nil {
			log.Printf("⚠️  Warning: topic not found: %s\n", entry.TopicID)
			continue
		}

		// Add only time as header2 (topic already added at the top)
		contentParts = append(contentParts, fmt.Sprintf("## %s", entryTime.Format("15:04")))
		contentParts = append(contentParts, "")

		// Upload image if present
		var imageURL string
		if entry.ImageID != nil && *entry.ImageID != "" {
			mediaPath, err := m.flowgresClient.GetMediaPath(*entry.ImageID)
			if err == nil {
				fullPath := filepath.Join(m.backupPath, mediaPath)
				resourceURL, err := m.memosClient.UploadResource(fullPath, m.dryRun)
				if err != nil {
					log.Printf("⚠️  Warning: Failed to upload image %s: %v\n", *entry.ImageID, err)
					*failedImages++
				} else {
					imageURL = resourceURL
					*uploadedImages++
				}
			}
		}

		// Add image if present
		if imageURL != "" {
			contentParts = append(contentParts, fmt.Sprintf("![image](%s)", imageURL))
			contentParts = append(contentParts, "")
		}

		// Add text content
		if entry.Text != nil && *entry.Text != "" {
			contentParts = append(contentParts, *entry.Text)
			contentParts = append(contentParts, "")
		}

		// Collect tags from this entry (categories only, skip Lifegoal)
		if topic != nil && topic.CategoryName != nil && *topic.CategoryName != "" {
			categoryName := *topic.CategoryName
			// Skip "Lifegoal" tag
			if categoryName != "Lifegoal" {
				categoryTag := strings.ReplaceAll(categoryName, " ", "-")
				categoryTag = strings.ReplaceAll(categoryTag, "#", "")
				allTags[fmt.Sprintf("#%s", categoryTag)] = true
			}
		}

		// Add separator between entries (except for the last one)
		contentParts = append(contentParts, "---")
		contentParts = append(contentParts, "")
	}

	// Remove the last separator
	if len(contentParts) >= 2 {
		contentParts = contentParts[:len(contentParts)-2]
	}

	// Add all collected tags at the end
	if len(allTags) > 0 {
		var tags []string
		for tag := range allTags {
			tags = append(tags, tag)
		}
		contentParts = append(contentParts, "")
		contentParts = append(contentParts, strings.Join(tags, " "))
	}

	content := strings.TrimSpace(strings.Join(contentParts, "\n"))

	// Create memo with the first entry's timestamp
	if err := m.memosClient.CreateMemo(content, createdTime, m.dryRun); err != nil {
		return fmt.Errorf("failed to create memo: %w", err)
	}

	return nil
}

// formatMemoContent formats an entry as Memos markdown content
func (m *Migrator) formatMemoContent(entry *flowgres.Entry, topic *flowgres.Topic, imageURL string) string {
	var lines []string

	// Add image if present
	if imageURL != "" {
		lines = append(lines, fmt.Sprintf("![image](%s)", imageURL))
		lines = append(lines, "")
	}

	// Add text content
	if entry.Text != nil && *entry.Text != "" {
		lines = append(lines, *entry.Text)
		lines = append(lines, "")
	}

	// Add tags
	var tags []string

	// Add topic as tag
	if topic != nil && topic.Name != "" {
		topicTag := strings.ReplaceAll(topic.Name, " ", "-")
		topicTag = strings.ReplaceAll(topicTag, "#", "")
		tags = append(tags, fmt.Sprintf("#%s", topicTag))
	}

	// Add category as tag
	if topic != nil && topic.CategoryName != nil && *topic.CategoryName != "" {
		categoryTag := strings.ReplaceAll(*topic.CategoryName, " ", "-")
		categoryTag = strings.ReplaceAll(categoryTag, "#", "")
		tags = append(tags, fmt.Sprintf("#%s", categoryTag))
	}

	if len(tags) > 0 {
		lines = append(lines, strings.Join(tags, " "))
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// parseTimestamp parses Flowgres timestamp formats
func parseTimestamp(timestamp string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestamp); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestamp)
}
