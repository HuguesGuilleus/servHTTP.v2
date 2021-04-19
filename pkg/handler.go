// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Add a file handler from file root r for the host h and the path p.
func (s *S) AddFile(h, p, r string, static *Static) error {
	if !path.IsAbs(r) {
		return fmt.Errorf("%q is not an absolute path", r)
	}
	d := http.Dir(r)

	s.AddHandler(h, p, "FILE", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, stat, err := openFile(d, r.URL.Path)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
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
				serverDir(w, r, f, static)
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

type File struct {
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modTime"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	HSize   string    `json:"hSize"`
}

func newFile(f os.FileInfo) File {
	nf := File{
		IsDir:   f.IsDir(),
		ModTime: f.ModTime(),
		Name:    f.Name(),
		Size:    f.Size(),
	}
	if s := nf.Size; s < 1000 {
		nf.HSize = fmt.Sprintf("%d\u00A0o", s)
	} else if s < 1000_000 {
		nf.HSize = fmt.Sprintf("%.1f\u00A0K", float64(s)/1000)
	} else if s < 1000_000_000 {
		nf.HSize = fmt.Sprintf("%.1f\u00A0M", float64(s)/1000_000)
	} else if s < 1000_000_000_000 {
		nf.HSize = fmt.Sprintf("%.1f\u00A0G", float64(s)/1000_000_000)
	} else {
		nf.HSize = fmt.Sprintf("%.1f\u00A0T", float64(s)/1000_000_000_000)
	}
	return nf
}

// Generate and index, serialise it or use template and send it to the client.
func serverDir(w http.ResponseWriter, r *http.Request, dir http.File, static *Static) {
	lister := func() []File {
		listInfo, _ := dir.Readdir(0)
		list := make([]File, 0, len(listInfo))
		for _, f := range listInfo {
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}
			list = append(list, newFile(f))
		}
		sort.Slice(list, func(i int, j int) bool {
			if list[i].IsDir != list[j].IsDir {
				return list[i].IsDir
			}
			return list[i].Name < list[j].Name
		})
		return list
	}

	stat, _ := dir.Stat()
	w.Header().Set("Date", stat.ModTime().Format(time.RFC1123))
	switch r.URL.Query().Get("f") {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		j, _ := json.Marshal(lister())
		w.Write(j)
	case "js":
		w.Header().Set("Content-Type", "text/html")
		w.Write(static.IndexJs())
		if p, ok := w.(http.Pusher); ok {
			u := *r.URL
			q := r.URL.Query()
			q.Set("f", "json")
			u.RawQuery = q.Encode()
			p.Push(u.String(), nil)
		}
	default:
		w.Header().Set("Content-Type", "text/html")
		static.Index().Execute(w, lister())
	}
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
		w.Write(static.E502())
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
			http.Redirect(w, r, to+r.RequestURI, http.StatusMovedPermanently)
		}
	}())
	return nil
}

// Add a hanlder for the host h, and with the path p. t is the name type of the
// handlern, it uses by the logger.
func (s *S) AddHandler(h, p, t string, f http.Handler) {
	if s.hosts == nil {
		s.hosts = make(map[string]bool)
	}
	s.hosts[h] = true

	prefix := "[" + t + "] " + h + p

	s.m.HandleFunc(h+p, func(w http.ResponseWriter, r *http.Request) {
		logReq(prefix, "", r)

		r2 := *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = "/" + strings.TrimPrefix(r2.URL.Path, p)

		f.ServeHTTP(w, &r2)
	})
}
