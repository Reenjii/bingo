package bingo

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// An entry in the paste index.
type indexEntry struct {
	id     string
	expire time.Time
}

// ByExpire implements sort.Interface for []indexEntry based on the expire field.
type ByExpire []indexEntry

func (a ByExpire) Len() int           { return len(a) }
func (a ByExpire) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByExpire) Less(i, j int) bool { return a[i].expire.Before(a[j].expire) }

// Paste index.
// RWMutex ensures safe concurrent access to the slice.
var index = struct {
	sync.RWMutex
	s []indexEntry
}{
	s: make([]indexEntry, 0, 10),
}

// Add a paste to the index.
func (paste *Paste) index() {
	index.Lock()
	index.s = append(index.s, indexEntry{paste.Id, paste.Expire})
	index.Unlock()
}

// Build the paste index from disk.
func buildIndex() error {
	Loggers.Info.Println("Build paste index...")
	e := indexFolder(conf.Root, "")
	Loggers.Info.Printf("Paste index built with %d entries", len(index.s))
	return e
}

// Scan a folder for paste indexing.
// This function is called recursively when a subfolder is encountered.
// prefix keps track of the folder hierarchy to rebuild pastes ids.
func indexFolder(path, prefix string) error {
	matches, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		stat, e := os.Stat(match)
		if e != nil {
			return e
		}

		if stat.IsDir() {
			// Recursively index folder
			// Unless this is a discussion folder (ends with _)
			if match[len(match)-1:] != "_" {
				indexFolder(match, prefix+filepath.Base(match))
			}
		}

		if stat.Mode().IsRegular() {
			// Regular file, load and index paste
			paste, loadError := loadPaste(prefix + filepath.Base(match))
			if loadError != nil {
				return loadError
			}
			paste.index()
		}
	}

	return nil
}

// Delete expired pastes from disk according to index data.
func deleteExpiredPastes() {
	Loggers.Info.Println("Delete expired pastes according to index data")

	// Sort the index by increasing expiration dates
	index.Lock()
	sort.Sort(ByExpire(index.s))
	index.Unlock()

	// Remove expired pastes (they are at the beginning, if any)
	max := -1
	for i, e := range index.s {
		if e.expire.Before(time.Now()) {
			// This paste has expired
			max = i
			paste, loadError := loadPaste(e.id)
			if loadError != nil {
				Loggers.Warn.Printf("Paste %s must be deleted (expired) but cannot be loaded (maybe already deleted ?): %s", e.id, loadError.Error())
			} else {
				if delError := paste.del(); delError != nil {
					Loggers.Error.Printf("Cannot delete expired paste %s: %s", e.id, delError.Error())
				}
			}
		} else {
			// We found a non-expired paste, stop here
			break
		}
	}

	// Remove deleted pastes from index
	if max >= 0 {
		index.Lock()
		index.s = index.s[max+1:]
		index.Unlock()
	}
}

// Start the clean daemon.
func startCleanDaemon() {
	Loggers.Info.Printf("Start clean daemon with a %d seconds threshold", conf.CleanThreshold)
	tick := time.NewTicker(time.Duration(conf.CleanThreshold) * time.Second).C
	go func() {
		for _ = range tick {
			deleteExpiredPastes()
		}
	}()
}

// Debug function that prints the index content.
func printIndex() {
	Loggers.Trace.Println("=== Index ===")
	for i, e := range index.s {
		Loggers.Trace.Printf("#%d %s at %s", i, e.id, e.expire)
	}
	Loggers.Trace.Println("=============")
}
