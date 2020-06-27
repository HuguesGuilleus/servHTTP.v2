// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Add a file handler from file root r for the host h and the path p.
func (s *S) AddFile(h, p, r string) error {
	s.AddHandler(h, p, "file", r, http.FileServer(http.Dir(r)))
	return nil
}

// Add a reverso proxy to the host
func (s *S) AddReverse(h, p, to string) error {
	u, err := url.Parse(to)
	if err != nil {
		return err
	}
	s.AddHandler(h, p, "reverse", to, httputil.NewSingleHostReverseProxy(u))
	return nil
}

// Add a retirect handler for the destination to.
func (s *S) AddRedirect(h, p, to string) error {
	s.AddHandler(h, p, "redirect", to, func() http.HandlerFunc {
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
func (s *S) AddHandler(h, p, t, param string, f http.Handler) {
	if s.hosts == nil {
		s.hosts = make(map[string]bool)
	}
	s.hosts[h] = true

	l := "[REQ]" + h + p + " " + t + "<" + param + ">"
	strip := http.StripPrefix(p, f).ServeHTTP

	s.m.HandleFunc(h+p, func(w http.ResponseWriter, r *http.Request) {
		log.Println(l, r.RemoteAddr, r.Method, maxURL(r.RequestURI))
		strip(w, r)
	})
}
