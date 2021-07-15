// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/template"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Simple server file from the root.
type FileServer struct {
	Base
	Root string
}

func (s *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.writeHeaders(w)

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		s.HTTP(logFileErrorMethod, r)
		template.ErrorString(w, r, http.StatusMethodNotAllowed, "Unsupported method.")
		return
	}

	p := filepath.Join(s.Root, filepath.FromSlash(s.Trim(r.URL.Path)))

	stat, err := os.Stat(p)
	if err != nil && os.IsNotExist(err) {
		s.HTTP(logFileErrorNotFound, r)
		template.ErrorString(w, r, http.StatusNotFound, "File not found.")
		return
	} else if err != nil {
		s.HTTP(logFileErrorInternal, r)
		template.ErrorString(w, r, http.StatusInternalServerError, internalReadError)
		return
	}
	w.Header().Set("Last-Modified", stat.ModTime().UTC().Format(http.TimeFormat))

	if stat.IsDir() {
		if !strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path += "/"
			http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		}
		list, err := ioutil.ReadDir(p)
		if err != nil {
			s.HTTP(logFileErrorInternal, r)
			template.ErrorString(w, r, http.StatusInternalServerError, internalReadError)
			return
		}
		if s.tryServeFile(w, r, filepath.Join(p, "index.html")) {
			s.HTTP(logFileIndex, r)
			template.Index(w, r.Host, r.URL.Path, list)
		}
	} else {
		if s.tryServeFile(w, r, p) {
			s.HTTP(logFileErrorInternal, r)
			template.ErrorString(w, r, http.StatusInternalServerError, internalReadError)
		}
	}
}

// Try to serve the file with path p, in compress mode or standard mode. Return true if fail.
func (s *FileServer) tryServeFile(w http.ResponseWriter, r *http.Request, p string) bool {
	e := r.Header.Get("Accept-Encoding")

	if strings.Contains(e, "gzip") {
		if f, err := os.Open(p + "$gzip"); err == nil {
			defer f.Close()
			s.HTTP(logFileCompressed, r)
			w.Header().Set("Content-Encoding", "gzip")
			if ext := mime.TypeByExtension(filepath.Ext(p)); ext != "" {
				w.Header().Set("Content-Type", ext)
			}
			io.Copy(w, f)
			return false
		}
	}

	if f, err := os.Open(p); err == nil {
		defer f.Close()
		s.HTTP(logFileServed, r)
		http.ServeContent(w, r, p, time.Time{}, f)
		return false
	}

	return true
}
