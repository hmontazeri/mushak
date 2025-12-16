package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/blang/semver"
	"github.com/hmontazeri/mushak/pkg/version"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const (
	updateCheckInterval = 24 * time.Hour
)

// CheckUpdateAsync starts the update check and returns a function to wait/print results
func CheckUpdateAsync() func() {
	if version.GetVersion() == "dev" {
		return func() {}
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return func() {}
	}
	checkFile := filepath.Join(cacheDir, "mushak", "last_update_check")

	// Check cache
	if info, err := os.Stat(checkFile); err == nil {
		if time.Since(info.ModTime()) < updateCheckInterval {
			return func() {}
		}
	}

	// Outcome channel
	type result struct {
		version string
		err     error
	}
	done := make(chan result, 1)

	go func() {
		// Update timestamp first
		os.MkdirAll(filepath.Dir(checkFile), 0755)
		os.WriteFile(checkFile, []byte(time.Now().String()), 0644)

		latest, found, err := selfupdate.DetectLatest("hmontazeri/mushak")
		if err != nil {
			done <- result{err: err}
			return
		}
		if !found {
			done <- result{err: fmt.Errorf("not found")}
			return
		}
		done <- result{version: latest.Version.String()}
	}()

	return func() {
		select {
		case res := <-done:
			if res.err == nil {
				current, _ := semver.Make(version.GetVersion())
				latest, _ := semver.Make(res.version)
				if latest.GT(current) {
					fmt.Fprintf(os.Stderr, "\n\nUpdate available %s -> %s\nRun 'mushak upgrade' to update\n", current, latest)
				}
			}
		case <-time.After(500 * time.Millisecond):
			// Don't wait too long at exit
			return
		}
	}
}
