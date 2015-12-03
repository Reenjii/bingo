package bingo

import (
	"crypto/sha1"
	"encoding/hex"
	"sync"
	"time"
)

// Antiflood map contains times at which an ip (hashed) posted data.
// RWMutex ensures safe concurrent access to the map.
var antiflood = struct {
	sync.RWMutex
	m map[string]time.Time
}{
	m: make(map[string]time.Time),
}

// Returns the sha1 hash of the input string.
func hash(ip string) string {
	hash := sha1.Sum([]byte(ip))
	return hex.EncodeToString(hash[:])
}

// Registers an ip in the flood map.
func updateFlood(ip string) {
	h := hash(ip)
	antiflood.Lock()
	antiflood.m[h] = time.Now()
	antiflood.Unlock()
}

// Check whether an ip is flooding.
func isFlood(ip string) bool {
	h := hash(ip)
	antiflood.RLock()
	d := time.Since(antiflood.m[h])
	antiflood.RUnlock()
	return d < time.Duration(conf.FloodThreshold)*time.Second
}
