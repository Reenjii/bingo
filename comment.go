package bingo

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"
)

/*
A comment.

 - Id: comment id
 - Author: comment author (encrypted)
 - Avatar: author avatar
 - Data: comment (encrypted) data
 - Postdate: comment creation date
 - Highlight: whether to enable syntax highlighting
 - Parent: parent comment, if any
*/
type Comment struct {
	Id        string    `json:"id"`
	Author    string    `json:"author"`
	Avatar    string    `json:"avatar"`
	Data      string    `json:"data"`
	Postdate  time.Time `json:"postdate"`
	Highlight bool      `json:"highlight"`
	Parent    string    `json:"parent"`
}

// CommentsByDate implements sort.Interface for []Comment based on the Postdate field.
type CommentsByDate []Comment

func (a CommentsByDate) Len() int           { return len(a) }
func (a CommentsByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommentsByDate) Less(i, j int) bool { return a[i].Postdate.Before(a[j].Postdate) }

// Create a new comment.
// Setup comment postdate and id.
func newComment(data string, parent *Comment) Comment {
	comment := Comment{
		Data:     data,
		Postdate: time.Now(),
	}
	if parent != nil {
		comment.Parent = parent.Id
	}
	comment.computeId()
	return comment
}

// Compute comment id.
func (comment *Comment) computeId() {
	hash := sha1.Sum([]byte(comment.Data))
	comment.Id = hex.EncodeToString(hash[:10])
}

// Compute avatar
func (comment *Comment) computeAvatar(ip string) {
	a := Avatar{X: 32, Y: 32}
	comment.Avatar = a.Avatar(ip)
}

// Compute the storage path of a comment.
func (comment *Comment) storagePath(paste *Paste) string {
	// Compute paste discussion folder
	s := paste.discussionPath()
	// Build comment path
	s = filepath.Clean(path.Join(s, comment.Id))
	Loggers.Info.Printf("Computed comment storage path %s", s)
	return s
}

// Save a comment to disk.
func (comment *Comment) save(paste *Paste) error {
	Loggers.Info.Printf("Save comment %s", comment.Id)

	p := comment.storagePath(paste)

	if err := setupFolder(filepath.Dir(p), 0770); err != nil {
		return err
	}

	// Marshal comment
	s, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(p, s, 0640)
}

// Load a comment from disk.
func loadComment(id string, paste *Paste) (Comment, error) {
	Loggers.Info.Printf("Load comment %s", id)

	comment := &Comment{Id: id}
	p := comment.storagePath(paste)

	// Read file
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return Comment{}, err
	}

	// Unmarshal data
	if err := json.Unmarshal(data, &comment); err != nil {
		return Comment{}, err
	}

	return *comment, nil
}
