// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	_ "embed"
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	//go:embed template.gohtml
	_tempSrcHtml string
	//go:embed template_player.js
	_tempSrcJs0 string
	_tempSrcJs1 = regexp.MustCompile(`\s+`).ReplaceAllString(_tempSrcJs0, " ")
	_tempSrcJs2 = regexp.MustCompile(`(\W) | (\W)`).ReplaceAllString(_tempSrcJs1, "$1$2")

	// The html template for error and index
	temp = template.Must(template.New("").Parse(
		strings.Replace(_tempSrcHtml, "/*PLAYER*/", _tempSrcJs2, 1),
	))
)

// Generated html index for host and path with the list of files.
func Index(w http.ResponseWriter, host, path string, files []fs.FileInfo) {
	p := page{Files: make([]FileInfo, 0, len(files))}
	p.setURL(host, path)
	numberDirectory := 0
	for _, f := range files {
		n := f.Name()
		var m *Media
		if strings.ContainsAny(n, "$") {
			continue
		} else if p.Readme == "" && (n == "README" || n == "README.txt" || n == "README.md") {
			p.Readme = n
			if p, ok := w.(http.Pusher); ok {
				p.Push(path+n, nil)
			}
		} else if isImage(n) {
			p.HasMedia = true
			m = &Media{Image: n}
		} else if isAudio(n) {
			p.HasMedia = true
			m = &Media{Audio: n}
		} else if isVideo(n) {
			p.HasMedia = true
			m = &Media{Video: n}
		}

		isDir := f.IsDir()
		if isDir {
			numberDirectory++
		}
		size := f.Size()
		p.Files = append(p.Files, FileInfo{
			Name:    n,
			Size:    size,
			SizeHum: humainSize(size),
			IsDir:   isDir,
			ModTime: f.ModTime().UTC().Truncate(time.Second),
			Media:   m,
		})
	}
	sort.Slice(p.Files, func(i int, j int) bool {
		if p.Files[i].IsDir != p.Files[j].IsDir {
			return p.Files[i].IsDir
		} else {
			return p.Files[i].Name < p.Files[j].Name
		}
	})
	// onlyFiles := p.Files[numberDirectory:]

	w.Header().Set("Content-Type", "text/html")

	temp.Execute(w, &p)
}

// isAudio return true if the file name has an audio extension.
func isAudio(n string) bool {
	return strings.HasSuffix(n, ".mp3") ||
		strings.HasSuffix(n, ".wav") ||
		strings.HasSuffix(n, ".flac") ||
		strings.HasSuffix(n, ".ogg")
}

// isImage return true if the file name has an image extension.
func isImage(n string) bool {
	return strings.HasSuffix(n, ".bmp") ||
		strings.HasSuffix(n, ".jpeg") ||
		strings.HasSuffix(n, ".jpg") ||
		strings.HasSuffix(n, ".png") ||
		strings.HasSuffix(n, ".webp")
}

// isVideo return true if the file has an video extension.
func isVideo(n string) bool {
	return strings.HasSuffix(n, ".mp4") ||
		strings.HasSuffix(n, ".webm") ||
		strings.HasSuffix(n, ".mpeg") ||
		strings.HasSuffix(n, ".png") ||
		strings.HasSuffix(n, ".webp")
}

// isVideo return true if the file has an text track extension.
func isTextTrack(n string) bool {
	return strings.HasSuffix(n, ".srt") || strings.HasSuffix(n, ".vtt")
}

// Response with the error.
func ErrorString(w http.ResponseWriter, r *http.Request, code int, err string) {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		p := page{
			ErrStatus: err,
			ErrCode:   code,
		}
		p.setURL(r.Host, r.URL.Path)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(code)
		temp.Execute(w, &p)
	} else {
		w.WriteHeader(code)
		w.Write([]byte("Error: "))
		w.Write([]byte(err))
		w.Write([]byte("\r\n"))
	}
}
