package bingo

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

/*
A paste.

 - Id: paste id
 - Data: paste (encrypted) data
 - Expire: paste expiration date
 - Postdate: paste creation date
 - Burn: whether this paste must be deleted once read
 - Highlight: whether to enable syntax highlighting
 - Discussion: whether discussions are enabled
 - Comments: paste comments
*/
type Paste struct {
	Id         string    `json:"id"`
	Data       string    `json:"data"`
	Expire     time.Time `json:"expire"`
	Postdate   time.Time `json:"postdate"`
	Burn       bool      `json:"burn"`
	Highlight  bool      `json:"highlight"`
	Discussion bool      `json:"discussion"`
	Comments   []Comment `json:"comments"`
}

// Create a new paste.
// Setup paste postdate and id.
func newPaste(data string) Paste {
	paste := Paste{
		Data:     data,
		Postdate: time.Now(),
	}
	paste.computeId()
	return paste
}

// Compute paste id from data.
func (paste *Paste) computeId() {
	hash := sha1.Sum([]byte(paste.Data))
	paste.Id = hex.EncodeToString(hash[:10])
}

// Compute a paste delete token.
func (paste *Paste) hmac(key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(paste.Data))
	hash := mac.Sum(nil)
	return hex.EncodeToString(hash[:10])
}

// Validate a paste delete token.
func (paste *Paste) hmacValidate(token string, key []byte) bool {
	expected, err := hex.DecodeString(token)
	if err != nil {
		Loggers.Error.Panicf("Cannot decode token %s: %s", token, err.Error())
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(paste.Data))
	computed := mac.Sum(nil)
	return hmac.Equal(computed[:10], expected)
}

// Check if a paste has expired.
func (paste *Paste) hasExpired() bool {
	return paste.Expire.Before(time.Now())
}

// Compute the storage path of a paste.
func (paste *Paste) storagePath() string {
	if 2*conf.Depth >= len(paste.Id) {
		panic(errors.New("depth too big"))
	}
	s := conf.Root
	for i := 0; i < conf.Depth; i++ {
		s = path.Join(s, paste.Id[2*i:2*(i+1)])
	}
	s = filepath.Clean(path.Join(s, paste.Id[2*conf.Depth:]))
	Loggers.Trace.Printf("Computed paste %s storage path: %s", paste.Id, s)
	return s
}

// Compute the discussion folder path of a paste.
func (paste *Paste) discussionPath() string {
	return paste.storagePath() + "_"
}

// Save a paste to disk.
func (paste *Paste) save() error {
	Loggers.Info.Printf("Save paste %s", paste.Id)

	p := paste.storagePath()

	if err := setupFolder(filepath.Dir(p), 0770); err != nil {
		return err
	}

	// Marshal paste
	s, err := json.Marshal(paste)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(p, s, 0640)
}

// Load a paste from disk.
func loadPaste(id string) (Paste, error) {
	Loggers.Info.Printf("Load paste %s", id)

	paste := &Paste{Id: id}
	p := paste.storagePath()

	// Read file
	data, err := ioutil.ReadFile(p)
	if err != nil {
		Loggers.Error.Printf("Paste read error %s: %s", id, err)
		return Paste{}, err
	}

	// Unmarshal data
	if err := json.Unmarshal(data, &paste); err != nil {
		Loggers.Error.Printf("Paste unmarshal error %s: %s", id, err)
		return Paste{}, err
	}

	return *paste, nil
}

// Delete a paste from disk.
func (paste *Paste) del() error {
	Loggers.Info.Printf("Delete paste %s", paste.Id)

	p := paste.storagePath()

	if err := os.Remove(p); err != nil {
		return err
	}

	// Delete paste discussion if any
	if paste.Discussion {
		d := paste.discussionPath()
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}

	return nil
}

// Load a paste comments.
func (paste *Paste) loadComments() error {
	Loggers.Info.Printf("Load comments of paste %s", paste.Id)

	p := paste.discussionPath()

	matches, err := filepath.Glob(filepath.Join(p, "*"))
	if err != nil {
		return err
	}

	paste.Comments = make([]Comment, len(matches))

	for i, s := range matches {
		c, err := loadComment(filepath.Base(s), paste)
		if err != nil {
			return err
		}
		paste.Comments[i] = c
	}

	sort.Sort(CommentsByDate(paste.Comments))

	return nil
}
