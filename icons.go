package main

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
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

func iconFor(de os.DirEntry) template.HTML {
	if de.IsDir() {
		return iconFolder
	}
	if ic, ok := iconByExt[strings.ToLower(filepath.Ext(de.Name()))]; ok {
		return ic
	}
	return iconDoc
}
