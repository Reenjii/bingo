package bingo

import (
	"testing"
)

func init() {
	// Disable logging
	setVerbosity(0)
}

func TestComputeId(t *testing.T) {

	pastes := []struct {
		data, id string
	}{
		{"Awesome paste", "d9441ab2ce8126457ecd"},
		{"1337", "77ba9cd915c8e359d973"},
		{"", "da39a3ee5e6b4b0d3255"},
	}

	for _, p := range pastes {
		paste := newPaste(p.data)
		if paste.Id != p.id {
			t.Errorf("newPaste(%q).Id == %q, want %q", p.data, paste.Id, p.id)
		}
	}

}

func TestPath(t *testing.T) {

	// Setup conf
	conf.Root = "/path/to/dir/"
	// With default depth (2)

	pastes := []struct {
		data, id, path string
	}{
		{"Awesome paste", "d9441ab2ce8126457ecd", conf.Root + "d9/44/1ab2ce8126457ecd"},
		{"1337", "77ba9cd915c8e359d973", conf.Root + "77/ba/9cd915c8e359d973"},
		{"", "da39a3ee5e6b4b0d3255", conf.Root + "da/39/a3ee5e6b4b0d3255"},
	}

	for _, p := range pastes {
		paste := newPaste(p.data)
		path := paste.storagePath()
		dpath := paste.discussionPath()
		if path != p.path {
			t.Errorf("newPaste(%q).storagePath() == %q, want %q", p.data, path, p.path)
		}
		if dpath != p.path+"_" {
			t.Errorf("newPaste(%q).discussionPath() == %q, want %q", p.data, path, p.path+"_")
		}
	}

	// With custom depth
	conf.Depth = 5

	pastes = []struct {
		data, id, path string
	}{
		{"Awesome paste", "d9441ab2ce8126457ecd", conf.Root + "d9/44/1a/b2/ce/8126457ecd"},
		{"1337", "77ba9cd915c8e359d973", conf.Root + "77/ba/9c/d9/15/c8e359d973"},
		{"", "da39a3ee5e6b4b0d3255", conf.Root + "da/39/a3/ee/5e/6b4b0d3255"},
	}

	for _, p := range pastes {
		paste := newPaste(p.data)
		path := paste.storagePath()
		dpath := paste.discussionPath()
		if path != p.path {
			t.Errorf("newPaste(%q).storagePath() == %q, want %q", p.data, path, p.path)
		}
		if dpath != p.path+"_" {
			t.Errorf("newPaste(%q).discussionPath() == %q, want %q", p.data, path, p.path+"_")
		}
	}

}

func TestDeleteToken(t *testing.T) {

	pastes := []struct {
		data, token string
	}{
		{"Awesome paste", "035fd1a9ccb554b8cb8f"},
		{"1337", "236f915ae883155a5766"},
		{"", "537c24565a8207e2b7d9"},
	}

	// Server secret key
	key := []byte("hakuna matata")

	for _, p := range pastes {
		paste := newPaste(p.data)
		hmac := paste.hmac(key)
		if hmac != p.token {
			t.Errorf("newPaste(%q).hmac(<key>) == %q, want %q", p.data, hmac, p.token)
		}
	}

	for _, p := range pastes {
		paste := newPaste(p.data)
		if !paste.hmacValidate(p.token, key) {
			t.Errorf("newPaste(%q).hmacValidate(%q, <key>) is false, want true", p.data, p.token)
		}
	}

}
