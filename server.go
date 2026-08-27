package main

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed src/html/*.html
var htmlFS embed.FS

//go:embed src/icons/*.svg
var iconFS embed.FS

var (
	indexTmpl    = template.Must(template.ParseFS(htmlFS, "src/html/indexof.html"))
	notFoundTmpl = template.Must(template.ParseFS(htmlFS, "src/html/notfound.html"))
)

var (
	iconFolder = mustIcon("folder-icon.svg")
	iconParent = mustIcon("arrow-return.svg")
	iconDoc    = mustIcon("docs-file.svg")
	iconZip    = mustIcon("folder-zip.svg")
	iconAudio  = mustIcon("audio-file.svg")
	iconVideo  = mustIcon("video-file.svg")
)

var iconByExt = map[string]template.HTML{
	".zip": iconZip, ".tar": iconZip, ".gz": iconZip, ".tgz": iconZip,
	".bz2": iconZip, ".xz": iconZip, ".rar": iconZip, ".7z": iconZip,

	".mp3": iconAudio, ".wav": iconAudio, ".flac": iconAudio,
	".ogg": iconAudio, ".m4a": iconAudio, ".aac": iconAudio, ".opus": iconAudio,

	".mp4": iconVideo, ".mkv": iconVideo, ".webm": iconVideo,
	".mov": iconVideo, ".avi": iconVideo, ".m4v": iconVideo,
}

// template.HTML skips escaping, which is safe only because these are our own
// embedded files and never anything from the served directory.
func mustIcon(name string) template.HTML {
	b, err := iconFS.ReadFile("src/icons/" + name)
	if err != nil {
		panic(err)
	}
	return template.HTML(b)
}

func iconFor(de os.DirEntry) template.HTML {
	if de.IsDir() {
		return iconFolder
	}
	if ic, ok := iconByExt[strings.ToLower(filepath.Ext(de.Name()))]; ok {
		return ic
	}
	return iconDoc
}

type entry struct {
	Name string
	URL  string
	Icon template.HTML
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

func (s *server) listDir(w http.ResponseWriter, upath, full string) {
	names, err := os.ReadDir(full)
	if err != nil {
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	entries := make([]entry, 0, len(names))
	for _, de := range names {
		link := url.URL{Path: de.Name()}
		e := entry{Name: de.Name(), URL: link.String(), Icon: iconFor(de)}
		if de.IsDir() {
			e.Name += "/"
			e.URL += "/"
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
