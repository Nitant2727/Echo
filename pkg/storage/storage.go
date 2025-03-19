package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Storage struct {
	filePath string
	mutex    sync.RWMutex
	data     *windowData
}

type WindowStats struct {
	Title    string
	Duration time.Duration
	Date     time.Time
}

type windowData struct {
	Stats []WindowStats `json:"stats"`
}

func NewStorage(filePath string) (*Storage, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	s := &Storage{
		filePath: filePath,
		data:     &windowData{Stats: []WindowStats{}},
	}

	if _, err := os.Stat(filePath); err == nil {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read storage file: %v", err)
		}
		if err := json.Unmarshal(file, s.data); err != nil {
			return nil, fmt.Errorf("failed to parse storage file: %v", err)
		}
	}

	return s, nil
}

func (s *Storage) Close() error {
	return s.save()
}

func (s *Storage) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Storage) SaveWindowStats(title string, duration time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data.Stats = append(s.data.Stats, WindowStats{
		Title:    title,
		Duration: duration,
		Date:     time.Now(),
	})

	return s.save()
}

func (s *Storage) GetDailyStats() ([]WindowStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	oneDayAgo := time.Now().Add(-24 * time.Hour)
	statsMap := make(map[string]*WindowStats)

	for _, stat := range s.data.Stats {
		if stat.Date.After(oneDayAgo) {
			if existing, ok := statsMap[stat.Title]; ok {
				existing.Duration += stat.Duration
				if stat.Date.After(existing.Date) {
					existing.Date = stat.Date
				}
			} else {
				statsMap[stat.Title] = &WindowStats{
					Title:    stat.Title,
					Duration: stat.Duration,
					Date:     stat.Date,
				}
			}
		}
	}

	var stats []WindowStats
	for _, stat := range statsMap {
		stats = append(stats, *stat)
	}

	return stats, nil
}