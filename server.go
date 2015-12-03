package bingo

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

/*
Postdata contains the json data sent by the client.

 - Data: paste (encrypted) data
 - Author: author (encrypted)
 - Expire: expiration date
 - Burn: whether this paste must be deleted once read
 - Highlight: whether to enable syntax highlighting
 - Discussion: whether discussions are enabled
 - Paste: parent paste, if any (for comments)
 - Parent: parent comment, if any (for comments)
 - Comments: whether this is a comment (true) or a regular paste (false)
*/
type Postdata struct {
	Data       string `json:"data"`
	Author     string `json:"author"`
	Expire     int    `json:"expire"`
	Burn       bool   `json:"burn"`
	Highlight  bool   `json:"highlight"`
	Discussion bool   `json:"discussion"`
	Paste      string `json="paste"`
	Parent     string `json="parent"`
	Comment    bool   `json="comment"`
}

/*
Postresponse contains the json response sent to the client.

 - Id: paste id
 - Postdate: paste creation date
 - Expire: expiration date
 - Delete: delete token
 - Avatar: author's avatar (comments only)
*/
type Postresponse struct {
	Id       string    `json:"id"`
	Postdate time.Time `json:"postdate"`
	Expire   time.Time `json:"expire"`
	Delete   string    `json:"delete"`
	Avatar   string    `json:"avatar"`
}

/*
Error response.

 - Code: error code
 - Error: error message
*/
type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

/*
Holds template data.

 - Paste: paste object
 - JPaste: marshaled paste
 - Deleted: true if the paste has been deleted
 - Code: error code
*/
type TemplateData struct {
	Paste   Paste
	JPaste  string
	Deleted bool
	Code    int
}

// Templates map.
var templates map[string]*template.Template

// URL patterns
var regexGetPaste *regexp.Regexp
var regexDeletePaste *regexp.Regexp

func init() {
	// Initialize URL patterns
	regexGetPaste = regexp.MustCompile("^/([A-Za-z0-9]{20})$")
	regexDeletePaste = regexp.MustCompile("^/delete/([A-Za-z0-9]{20})/([A-Za-z0-9]{20})$")
}

// Load templates on program initialisation
func initTemplates() {
	Loggers.Info.Println("Templates initialization")
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	views, err := filepath.Glob(conf.Views + "/*.html")
	if err != nil {
		Loggers.Error.Fatal(err)
	}
	Loggers.Info.Println("Views are in ", conf.Views+"/*.html")

	for _, view := range views {
		Loggers.Info.Println("Loading template", view)
		templates[filepath.Base(view)] = template.Must(template.ParseFiles(view))
	}

}

// renderTemplate is a wrapper around template.ExecuteTemplate.
func renderTemplate(w http.ResponseWriter, name string, data TemplateData) error {
	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		Loggers.Error.Printf("Template %s not found", name)
		return fmt.Errorf("The template %s does not exist.", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

func render(w http.ResponseWriter, data TemplateData) error {
	return renderTemplate(w, "paste.html", data)
}

func renderError(w http.ResponseWriter, code int, message string) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return renderTemplate(w, "paste.html", TemplateData{Code: code})
}

func renderAjaxError(w http.ResponseWriter, status int, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	response, marshalErr := json.Marshal(ErrorResponse{
		Code:  code,
		Error: message,
	})
	if marshalErr != nil {
		panic(marshalErr)
	}
	w.Write([]byte(response))
}

// Reads client's IP address from request data.
// IP address is read through RemoteAddr first.
// Fallbacks to X-Forwarded-For header when a local IP address is found in RemoteAddr.
// TODO support ipv6
func getIP(r *http.Request) string {
	// RemoteAddr has format IP:port
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if ip == "127.0.0.1" {
		// This is a local ip, use x-forwarded-for header instead
		forwardedFor := r.Header.Get("x-forwarded-for")
		if len(forwardedFor) > 0 {
			ip = forwardedFor
		}
	}
	return ip
}

// Handle root requests
func handlerRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if regexGetPaste.MatchString(r.URL.Path) {
			// Client wants to load a paste

			// Extract paste id from URL
			id := regexGetPaste.FindStringSubmatch(r.URL.Path)

			// Load paste from disk
			paste, err := loadPaste(id[1])
			if err != nil {
				renderError(w, 404, "Not found")
				return
			}

			// Map to store data for the template
			data := TemplateData{}

			// Has this paste expired ?
			if paste.hasExpired() {
				Loggers.Info.Printf("Paste %s has expired on %s, delete", paste.Id, paste.Expire)
				if err := paste.del(); err != nil {
					Loggers.Error.Printf("Cannot delete paste %s: %s", paste.Id, err)
				}

				// Return a 404 response
				renderError(w, 404, "Not found")
				return
			}

			// Should this paste be deleted after reading ?
			if paste.Burn {
				Loggers.Info.Printf("Burn paste %s after reading", paste.Id)
				if err := paste.del(); err != nil {
					Loggers.Error.Printf("Cannot delete paste %s: %s", paste.Id, err)
				}
			}

			// If paste discussion is enabled, load comments
			if paste.Discussion {
				if err := paste.loadComments(); err != nil {
					Loggers.Error.Printf("Cannot load comments of paste %s: %s", paste.Id, err)
					// TODO error
				}

				// Marshall comments
				cc := make([]string, len(paste.Comments))
				for i, comment := range paste.Comments {
					j, err := json.Marshal(comment)
					if err != nil {
						Loggers.Error.Printf("Cannot marshal comments of paste %s: %s", paste.Id, err)
						// TODO error
					}
					cc[i] = string(j)
				}
			}

			// Marshall paste
			jpaste, marshalErr := json.Marshal(paste)
			if marshalErr != nil {
				Loggers.Error.Printf("Cannot marshal paste %s: %s", paste.Id, marshalErr)
				renderError(w, 500, "Marshal error")
				return
			}

			data.JPaste = string(jpaste)
			data.Paste = paste

			if renderErr := render(w, data); renderErr != nil {
				Loggers.Error.Printf("Cannot render template for paste %s: %s", paste.Id, renderErr)
				renderError(w, 500, "Render error")
			}

			return
		} else if regexDeletePaste.MatchString(r.URL.Path) {
			// Client wants to delete a paste

			// Extract paste id and delete token from URL
			match := regexDeletePaste.FindStringSubmatch(r.URL.Path)
			id, token := match[1], match[2]
			Loggers.Info.Println("Delete paste", id, token)

			// Load paste from disk
			paste, err := loadPaste(id)
			if err != nil {
				renderError(w, 404, "Not found")
				return
			}

			// Validate delete token
			// TODO server's secret
			if !paste.hmacValidate(token, []byte("secret")) {
				Loggers.Warn.Println("Cannot validate token", token)
				renderError(w, 403, "Wrong delete token")
				return
			}

			if deleteErr := paste.del(); err != nil {
				Loggers.Error.Printf("Cannot delete paste %s: %s", paste.Id, deleteErr)
				renderError(w, 500, "Delete error")
				return
			}

			if renderErr := render(w, TemplateData{Deleted: true}); renderErr != nil {
				Loggers.Error.Printf("Cannot render template for paste %s: %s", paste.Id, renderErr)
				renderError(w, 500, "Render error")
				return
			}

			return
		} else {
			// Homepage
			if renderErr := render(w, TemplateData{}); renderErr != nil {
				Loggers.Error.Printf("Cannot render template: %s", renderErr)
				renderError(w, 500, "Render error")
			}

			return
		}
	case "POST":
		// Client wants to post a paste or a comment

		// Content-Type must be application/json
		contentType := r.Header.Get("Content-Type")
		matchJson, contentTypeErr := regexp.MatchString("application/json.*", contentType)
		// TODO support charset

		if contentTypeErr != nil {
			Loggers.Warn.Println("Cannot detect Content-Type", contentTypeErr)
			renderAjaxError(w, http.StatusBadRequest, http.StatusBadRequest, "Cannot detect content-type")
			return
		}

		if !matchJson {
			Loggers.Warn.Println("Wrong Content-Type, expecting application/json, got", contentType)
			renderAjaxError(w, http.StatusBadRequest, http.StatusBadRequest, "Wrong content-type")
			return
		}

		// Parse body
		decoder := json.NewDecoder(r.Body)
		var data Postdata
		decodeErr := decoder.Decode(&data)
		if decodeErr != nil {
			Loggers.Error.Printf("Cannot parse json data: %s", decodeErr)
			renderAjaxError(w, http.StatusBadRequest, http.StatusBadRequest, "Cannot parse request body")
			return
		}

		if data.Comment {
			// This is a comment

			// Check that paste exists
			paste, err := loadPaste(data.Paste)
			if err != nil {
				Loggers.Error.Printf("Cannot load paste %s: %s", data.Paste, err)
				renderAjaxError(w, http.StatusNotFound, http.StatusNotFound, "Paste not found")
				return
			}

			// Is discussion enabled ?
			if !paste.Discussion {
				Loggers.Error.Printf("Discussion is disabled for paste %s", data.Paste)
				renderAjaxError(w, http.StatusForbidden, http.StatusForbidden, "Discussion is disabled")
				return
			}

			// Handle parent comment, if any
			var parent *Comment = nil
			if data.Parent != "" {
				c, err := loadComment(data.Parent, &paste)
				if err != nil {
					Loggers.Error.Printf("Cannot load parent comment %s: %s", data.Parent, err)
					renderAjaxError(w, http.StatusNotFound, http.StatusNotFound, "Parent not found")
					return
				}
				parent = &c
			}

			// Check that user is not flooding
			if isFlood(getIP(r)) {
				Loggers.Warn.Println("Flood deteted")
				renderAjaxError(w, http.StatusForbidden, http.StatusForbidden, "Please wait before posting again")
				return
			}

			comment := newComment(data.Data, parent)
			comment.Highlight = data.Highlight
			comment.Author = data.Author
			comment.computeAvatar(getIP(r))

			if err := comment.save(&paste); err != nil {
				Loggers.Error.Printf("Unable to save comment %s: %s", comment.Id, err)
				renderAjaxError(w, http.StatusInternalServerError, http.StatusInternalServerError, "Could not save comment")
				return
			}

			// Update antiflood
			updateFlood(getIP(r))

			// Marshal response
			j, err := json.Marshal(Postresponse{
				Id:       comment.Id,
				Postdate: comment.Postdate,
				Delete:   "", // A comment cannot be deleted, the paste gets deleted
				Avatar:   comment.Avatar,
			})
			if err != nil {
				Loggers.Error.Printf("Marshal error: %s", err)
				renderAjaxError(w, http.StatusInternalServerError, http.StatusInternalServerError, "Marshal error")
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintf(w, "%s", j)

		} else {
			// This is a regular paste

			// Check that user is not flooding
			if isFlood(getIP(r)) {
				Loggers.Warn.Println("Flood deteted")
				renderAjaxError(w, http.StatusForbidden, http.StatusForbidden, "Please wait before posting again")
				return
			}

			p := newPaste(data.Data)
			p.Burn = data.Burn
			p.Discussion = data.Discussion
			p.Highlight = data.Highlight
			p.Expire = p.Postdate.Add(time.Duration(data.Expire) * time.Second)

			// TODO server's secret
			Loggers.Info.Println("Delete token is ", p.hmac([]byte("secret")))

			if err := p.save(); err != nil {
				Loggers.Error.Printf("Unable to save paste %s: %s", p.Id, err)
				renderAjaxError(w, http.StatusInternalServerError, http.StatusInternalServerError, "Could not save paste")
				return
			}

			// Update antiflood
			updateFlood(getIP(r))

			// Update index
			p.index()

			// Marshal response
			j, err := json.Marshal(Postresponse{
				Id:       p.Id,
				Postdate: p.Postdate,
				Expire:   p.Expire,
				Delete:   p.hmac([]byte("secret")),
			})
			if err != nil {
				Loggers.Error.Printf("Marshal error: %s", err)
				renderAjaxError(w, http.StatusInternalServerError, http.StatusInternalServerError, "Marshal error")
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintf(w, "%s", j)
		}
	}

}

func Serve(file string) {
	// Load configuration
	if err := conf.load(file); err != nil {
		panic(err)
	}

	// Initialize logging
	if conf.Log != "" {
		setupFolder(filepath.Dir(conf.Log), 0750)
		logfile, err := os.OpenFile(conf.Log, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			Loggers.Error.Printf("Error opening file: %v", err)
		} else {
			// Log file opened successfully
			if conf.Stdout {
				// Also log in stdout
				setAllOutput(io.MultiWriter(os.Stdout, logfile))
			} else {
				// Only log in logfile
				setAllOutput(logfile)
			}
		}
	}
	setVerbosity(conf.Verbosity)

	// Initialize templates
	initTemplates()

	// Initialize data folder
	if err := setupFolder(conf.Root, 0750); err != nil {
		panic(err)
	}

	// Build paste index
	buildIndex()

	// Start cleaner daemon
	startCleanDaemon()

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(conf.Static))))

	// Handle root
	http.HandleFunc("/", handlerRoot)

	addr := fmt.Sprintf(":%d", conf.Port)
	Loggers.Info.Println("Listening on", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
