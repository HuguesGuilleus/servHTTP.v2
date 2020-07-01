// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Add a file handler from file root r for the host h and the path p.
func (s *S) AddFile(h, p, r string, static *Static) error {

	if !path.IsAbs(r) {
		return fmt.Errorf("%q is not an absolute path", r)
	}
	d := http.Dir(r)

	// s.AddHandler(h, p, "FILE", http.FileServer(http.Dir(r)))
	s.AddHandler(h, p, "FILE", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, stat, err := openFile(d, r.URL.Path)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.Write(static.E404())
			return
		}
		defer f.Close()
		name := path.Base(r.URL.Path)

		if stat.IsDir() {
			if !strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusPermanentRedirect)
				return
			}
			index, statIndex, err := openFile(d, r.URL.Path+"index.html")
			if err != nil || statIndex.IsDir() {
				w.Header().Set("Content-Type", "text/html")
				list, _ := f.Readdir(0)
				static.Index().Execute(w, list)
				return
			}
			defer index.Close()
			f = index
			name = "index.html"
		}

		http.ServeContent(w, r, name, stat.ModTime(), f)
	}))
	return nil
}

// Open a http.File and its metat data.
func openFile(d http.Dir, p string) (http.File, os.FileInfo, error) {
	f, err := d.Open(filepath.FromSlash(p))
	if err != nil {
		return nil, nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	return f, stat, nil
}

// Add a reverso proxy to the host
func (s *S) AddReverse(h, p, to string, static *Static) error {
	u, err := url.Parse(to)
	if err != nil {
		return err
	}
	reverse := httputil.NewSingleHostReverseProxy(u)
	reverse.ErrorHandler = func(w http.ResponseWriter, r *http.Request, _ error) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		w.Write(static.E404())
	}
	reverse.ErrorLog = log.New(log.Writer(),
		"[REVERSE ERROR]",
		log.Flags()|log.Lmsgprefix)
	s.AddHandler(h, p, "REVERSE", reverse)
	return nil
}

// Add a retirect handler for the destination to.
func (s *S) AddRedirect(h, p, to string, _ *Static) error {
	s.AddHandler(h, p, "REDIRECT", func() http.HandlerFunc {
		if to[len(to)-1] == '/' {
			to = to[:len(to)-1]
		}
		return func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, to+r.URL.String(), http.StatusMovedPermanently)
		}
	}())
	return nil
}

// Add a hanlder for th host h, and with the path p.
func (s *S) AddHandler(h, p, t string, f http.Handler) {
	if s.hosts == nil {
		s.hosts = make(map[string]bool)
	}
	s.hosts[h] = true

	prefix := "[" + t + "] " + h + p
	strip := http.StripPrefix(p, f).ServeHTTP

	s.m.HandleFunc(h+p, func(w http.ResponseWriter, r *http.Request) {
		logReq(prefix, "", r)
		strip(w, r)
	})
}
