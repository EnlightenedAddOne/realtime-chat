package service

import (
	"os"
	"path/filepath"
	"time"

	"go-im-server/config"
	"go-im-server/pkg/logger"
)

// StartUploadCleanupWorker periodically deletes old files in upload directory.
func StartUploadCleanupWorker() {
	u := config.App.Upload
	if !u.CleanupEnabled {
		logger.Info.Println("upload cleanup worker disabled")
		return
	}

	interval := time.Duration(u.CleanupIntervalHours) * time.Hour
	maxAge := time.Duration(u.CleanupMaxAgeHours) * time.Hour
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	if maxAge <= 0 {
		maxAge = 7 * 24 * time.Hour
	}

	go func() {
		logger.Infof("upload cleanup worker started, interval=%s maxAge=%s", interval, maxAge)
		runCleanup(u.Path, maxAge)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runCleanup(u.Path, maxAge)
		}
	}()
}

func runCleanup(uploadPath string, maxAge time.Duration) {
	entries, err := os.ReadDir(uploadPath)
	if err != nil {
		logger.Error.Printf("upload cleanup read dir failed: %v", err)
		return
	}

	now := time.Now()
	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(uploadPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) < maxAge {
			continue
		}

		if err := os.Remove(filePath); err != nil {
			logger.Error.Printf("upload cleanup remove failed: %s err=%v", filePath, err)
			continue
		}
		removed++
	}

	if removed > 0 {
		logger.Infof("upload cleanup removed %d expired files", removed)
	}
}
