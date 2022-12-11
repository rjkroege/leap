package base

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SetupLogging configures the Go logger to write log messages to the "standard"
// place on MacOS. Logs for what.
// TODO(rjk): Generalize to more than Darwin
func SetupLogging(what string) error {
	// Based on how Kopia organizes its logs.
	logFileName := fmt.Sprintf("%v-%v-%v%v", what, time.Now().Format("20060102-150405"), os.Getpid(), ".log")

	logDir := filepath.Join(os.Getenv("HOME"), "Library", "Logs", what)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("can't make log directory %s: %v", logDir, err)
	}

	path := filepath.Join(logDir, logFileName)
	fd, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("can't create log file %s: %v", path, err)
	}

	log.SetOutput(fd)
	return nil
}

const oneday = time.Hour * 24

// RollLogs will strip old logs for what.
func RollLogs(what string) {
	if err := rollOneLog(what, oneday); err != nil {
		log.Printf("failed to roll %s logs: %v", what, err)
	}
}

func rollOneLog(target string, older time.Duration) error {
	now := time.Now()
	logDir := filepath.Join(os.Getenv("HOME"), "Library", "Logs", target)
	if err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(info.Name(), target) {
			return nil
		}
		if info.Mode().IsRegular() && now.Sub(info.ModTime()) > older {
			if err := os.Remove(path); err != nil {
				// This isn't fatal -- we just log this.
				log.Printf("can't delete old log message %q: %v", path, err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("rollOnLog %q: %v", logDir, err)
	}
	return nil
}
