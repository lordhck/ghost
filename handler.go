package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type entry struct {
	Name     string
	URL      string
	Icon     template.HTML
	Size     string
	Modified string
}

type indexData struct {
	Path       string
	ParentIcon template.HTML
	Entries    []entry
}

type server struct {
	dir string
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean() with a leading slash strips any ../ before it reaches the disk.
	upath := path.Clean("/" + r.URL.Path)
	full := filepath.Join(s.dir, filepath.FromSlash(upath))

	if isHidden(upath) {
		s.notFound(w, upath)
		return
	}

	info, err := os.Stat(full)
	if err != nil {
		s.notFound(w, upath)
		return
	}

	if !info.IsDir() {
		http.ServeFile(w, r, full)
		return
	}

	// Without the trailing slash the listing's relative links resolve one level too high.
	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, upath+"/", http.StatusMovedPermanently)
		return
	}

	index := filepath.Join(full, "index.html")
	if _, err := os.Stat(index); err != nil {
		s.listDir(w, upath, full)
		return
	}

	http.ServeFile(w, r, index)
}

// isHidden reports whether any component of upath is a dotfile, so that
// /.env is refused and not just omitted from the listing.
func isHidden(upath string) bool {
	for _, part := range strings.Split(upath, "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func (s *server) listDir(w http.ResponseWriter, upath, full string) {
	names, err := os.ReadDir(full)
	if err != nil {
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	entries := make([]entry, 0, len(names))
	for _, de := range names {
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}

		info, err := de.Info()
		if err != nil {
			continue
		}

		link := url.URL{Path: de.Name()}
		e := entry{
			Name:     de.Name(),
			URL:      link.String(),
			Icon:     iconFor(de),
			Size:     humanSize(info.Size()),
			Modified: info.ModTime().Format("2006-01-02 15:04"),
		}
		if de.IsDir() {
			e.Name += "/"
			e.URL += "/"
			e.Size = "-"
		}
		entries = append(entries, e)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexTmpl.Execute(w, indexData{Path: upath, ParentIcon: iconParent, Entries: entries})
}

func (s *server) notFound(w http.ResponseWriter, upath string) {
	custom := filepath.Join(s.dir, "404.html")
	data, err := os.ReadFile(custom)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	if err != nil {
		notFoundTmpl.Execute(w, map[string]string{"Path": upath})
		return
	}

	w.Write(data)
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
