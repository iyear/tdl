package persistence

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type TelegramDB struct {
	db   *sql.DB
	mu   sync.Mutex
	path string
}

func NewTelegramDB(path string) (*TelegramDB, error) {
	if path == "" {
		return nil, nil // Return nil if no database path is provided
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS telegram (
			channel_id INTEGER NOT NULL,
			message_id INTEGER NOT NULL,
			PRIMARY KEY (channel_id, message_id)
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &TelegramDB{
		db:   db,
		path: path,
	}, nil
}

func (t *TelegramDB) Close() error {
	if t == nil || t.db == nil {
		return nil
	}
	return t.db.Close()
}

func (t *TelegramDB) MessageExists(channelID int64, messageID int) (bool, error) {
	if t == nil || t.db == nil {
		return false, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	var exists int
	err := t.db.QueryRow("SELECT 1 FROM telegram WHERE channel_id = ? AND message_id = ?", channelID, messageID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check if message exists: %w", err)
	}
	return true, nil
}

func (t *TelegramDB) InsertMessage(channelID int64, messageID int) error {
	if t == nil || t.db == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, err := t.db.Exec("INSERT OR IGNORE INTO telegram (channel_id, message_id) VALUES (?, ?)", channelID, messageID)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}
