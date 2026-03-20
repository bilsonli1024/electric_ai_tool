package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogRotator struct {
	file         *os.File
	maxSize      int64
	maxAge       time.Duration
	currentSize  int64
	logDir       string
}

func NewLogRotator(logDir string, maxSizeMB int, maxAgeDays int) (*LogRotator, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	lr := &LogRotator{
		logDir:  logDir,
		maxSize: int64(maxSizeMB) * 1024 * 1024,
		maxAge:  time.Duration(maxAgeDays) * 24 * time.Hour,
	}

	if err := lr.openNewFile(); err != nil {
		return nil, err
	}

	go lr.cleanOldLogs()

	return lr, nil
}

func (lr *LogRotator) Write(p []byte) (n int, err error) {
	lr.currentSize += int64(len(p))
	
	if lr.currentSize >= lr.maxSize {
		lr.rotate()
	}

	return lr.file.Write(p)
}

func (lr *LogRotator) openNewFile() error {
	filename := filepath.Join(lr.logDir, time.Now().Format("2006-01-02_15-04-05")+".log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if lr.file != nil {
		lr.file.Close()
	}

	lr.file = file
	lr.currentSize = 0

	stat, err := file.Stat()
	if err == nil {
		lr.currentSize = stat.Size()
	}

	return nil
}

func (lr *LogRotator) rotate() error {
	return lr.openNewFile()
}

func (lr *LogRotator) cleanOldLogs() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		files, err := os.ReadDir(lr.logDir)
		if err != nil {
			continue
		}

		cutoff := time.Now().Add(-lr.maxAge)
		
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				os.Remove(filepath.Join(lr.logDir, file.Name()))
				log.Printf("🗑️  Removed old log file: %s", file.Name())
			}
		}
	}
}

func InitLogger(logDir string) error {
	rotator, err := NewLogRotator(logDir, 100, 7)
	if err != nil {
		return err
	}

	multiWriter := io.MultiWriter(os.Stdout, rotator)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	log.Println("📋 Logger initialized with rotation (max: 100MB, retention: 7 days)")
	return nil
}
