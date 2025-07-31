package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Backend implements datastorage.Backend for filesystem storage
type Backend struct {
	basePath string
	mu       sync.RWMutex
	closed   bool
}

// NewBackend creates a new filesystem data storage backend
func NewBackend(basePath string) (*Backend, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Backend{
		basePath: basePath,
	}, nil
}

// StoreData stores message data and returns the actual size written
func (b *Backend) StoreData(ctx context.Context, messageID string, data io.Reader) (int64, error) {
	if err := b.validateMessageID(messageID); err != nil {
		return 0, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, fmt.Errorf("backend is closed")
	}

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	dataPath := b.getDataPath(messageID)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(dataPath), 0o755); err != nil {
		return 0, fmt.Errorf("failed to create data directory: %w", err)
	}

	dataFile, err := os.Create(dataPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create data file: %w", err)
	}
	defer dataFile.Close()

	size, err := io.Copy(dataFile, data)
	if err != nil {
		return 0, fmt.Errorf("failed to write data: %w", err)
	}

	return size, nil
}

// GetDataReader returns a reader for message data
func (b *Backend) GetDataReader(ctx context.Context, messageID string) (io.ReadCloser, error) {
	if err := b.validateMessageID(messageID); err != nil {
		return nil, err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	dataPath := b.getDataPath(messageID)
	file, err := os.Open(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("message data not found: %s", messageID)
		}
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}

	return file, nil
}

// GetDataWriter returns a writer for message data
func (b *Backend) GetDataWriter(ctx context.Context, messageID string) (io.WriteCloser, error) {
	if err := b.validateMessageID(messageID); err != nil {
		return nil, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	dataPath := b.getDataPath(messageID)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(dataPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	file, err := os.Create(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create data file: %w", err)
	}

	return file, nil
}

// DeleteData removes message data
func (b *Backend) DeleteData(ctx context.Context, messageID string) error {
	if err := b.validateMessageID(messageID); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("backend is closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dataPath := b.getDataPath(messageID)
	err := os.Remove(dataPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete data file: %w", err)
	}

	// Clean up empty directories
	b.cleanupEmptyDirs(filepath.Dir(dataPath))

	return nil
}

// Close closes the data storage backend
func (b *Backend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	return nil
}

// Helper methods

func (b *Backend) validateMessageID(messageID string) error {
	if messageID == "" {
		return fmt.Errorf("messageID cannot be empty")
	}
	if strings.Contains(messageID, "..") {
		return fmt.Errorf("messageID cannot contain '..'")
	}
	if strings.ContainsAny(messageID, "/\\") {
		return fmt.Errorf("messageID cannot contain path separators")
	}
	if len(messageID) > 255 {
		return fmt.Errorf("messageID too long (max 255 characters)")
	}
	return nil
}

func (b *Backend) getDataPath(messageID string) string {
	// Use first 2 characters for directory sharding to avoid too many files in one directory
	if len(messageID) >= 2 {
		return filepath.Join(b.basePath, messageID[:2], messageID+".data")
	}
	return filepath.Join(b.basePath, "misc", messageID+".data")
}

func (b *Backend) cleanupEmptyDirs(dir string) {
	// Don't remove the base path
	if dir == b.basePath {
		return
	}

	// Check if directory is empty
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}

	// Remove empty directory
	if err := os.Remove(dir); err == nil {
		// Recursively clean up parent directory
		b.cleanupEmptyDirs(filepath.Dir(dir))
	}
}
