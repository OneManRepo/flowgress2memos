package flowgres

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Client handles interactions with Flowgres SQLite database
type Client struct {
	db         *sql.DB
	backupPath string
}

// NewClient creates a new Flowgres client
func NewClient(backupPath string) (*Client, error) {
	dbPath := filepath.Join(backupPath, "databases", "local_cordova_db.sqlite")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Client{
		db:         db,
		backupPath: backupPath,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// Entry represents a journal entry from Flowgres
type Entry struct {
	ID         string
	CreatedAt  string
	UpdatedAt  string
	TopicID    string
	MetadataID *string
	Type       string
	ImageID    *string
	VideoID    *string
	Text       *string
	IsVisible  bool
}

// Topic represents a topic/project in Flowgres
type Topic struct {
	ID           string
	Name         string
	CategoryID   *string
	CategoryName *string
}

// Category represents a category in Flowgres
type Category struct {
	ID   string
	Name string
}

// Media represents a media file reference
type Media struct {
	ID   string
	Path string
	Type string
}

// Location represents GPS coordinates
type Location struct {
	ID   string
	Lat  float64
	Lng  float64
	Name *string
}

// Metadata represents entry metadata
type Metadata struct {
	ID             string
	LocationID     *string
	MediaCreatedAt *string
}

// GetCategories retrieves all categories
func (c *Client) GetCategories() (map[string]string, error) {
	rows, err := c.db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	categories := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories[id] = name
	}

	return categories, nil
}

// GetTopics retrieves all topics with their category information
func (c *Client) GetTopics() (map[string]*Topic, error) {
	query := `
		SELECT t.id, t.name, t.category_id, c.name as category_name
		FROM topics t
		LEFT JOIN categories c ON t.category_id = c.id
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query topics: %w", err)
	}
	defer rows.Close()

	topics := make(map[string]*Topic)
	for rows.Next() {
		topic := &Topic{}
		if err := rows.Scan(&topic.ID, &topic.Name, &topic.CategoryID, &topic.CategoryName); err != nil {
			return nil, fmt.Errorf("failed to scan topic: %w", err)
		}
		topics[topic.ID] = topic
	}

	return topics, nil
}

// GetAllEntries retrieves all visible entries
func (c *Client) GetAllEntries() ([]*Entry, error) {
	query := `
		SELECT id, created_at, updated_at, topic_id, metadata_id,
		       type, image_id, video_id, text, is_visible
		FROM entries
		WHERE is_visible = 1
		ORDER BY created_at ASC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		entry := &Entry{}
		if err := rows.Scan(
			&entry.ID, &entry.CreatedAt, &entry.UpdatedAt, &entry.TopicID,
			&entry.MetadataID, &entry.Type, &entry.ImageID, &entry.VideoID,
			&entry.Text, &entry.IsVisible,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetMediaPath retrieves the file path for a media ID
func (c *Client) GetMediaPath(mediaID string) (string, error) {
	var path string
	err := c.db.QueryRow("SELECT path FROM media WHERE id = ?", mediaID).Scan(&path)
	if err != nil {
		return "", fmt.Errorf("failed to get media path: %w", err)
	}
	return path, nil
}

// GetLocation retrieves location information
func (c *Client) GetLocation(locationID string) (*Location, error) {
	location := &Location{}
	err := c.db.QueryRow(
		"SELECT id, lat, lng, name FROM locations WHERE id = ?",
		locationID,
	).Scan(&location.ID, &location.Lat, &location.Lng, &location.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	return location, nil
}

// GetMetadata retrieves metadata for an entry
func (c *Client) GetMetadata(metadataID string) (*Metadata, error) {
	metadata := &Metadata{}
	err := c.db.QueryRow(
		"SELECT id, location_id, media_created_at FROM metadata WHERE id = ?",
		metadataID,
	).Scan(&metadata.ID, &metadata.LocationID, &metadata.MediaCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	return metadata, nil
}

// GetBackupPath returns the backup directory path
func (c *Client) GetBackupPath() string {
	return c.backupPath
}
