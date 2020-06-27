// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
)

// One serve.
type S struct {
	m      http.ServeMux
	Config tls.Config
	hosts  map[string]bool
}

// Add a new certificat to the Configuration.
func (s *S) AddCert(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	s.Config.Certificates = append(s.Config.Certificates, cert)
	return nil
}

func (s *S) ServeTLS(a string) {
	if s.Config.NextProtos == nil {
		s.Config.NextProtos = []string{"h2", "http/1.1"}
	}

	log.Println("[LISTEN TLS]", a)

	l, err := tls.Listen("tcp", a, &s.Config)
	if err != nil {
		log.Fatal("[LISTEN ERROR]", a, err)
		return
	}
	http.Serve(l, &s.m)
}

// Serve to redirect all hosts to https
func (s *S) Serve(a, chalenge string) {
	log.Println("[LISTEN UNSAFE]", a, chalenge)
	err := http.ListenAndServe(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to response with a chalenge.
		if f, ok := openChalenge(chalenge, r.URL.Path); ok {
			log.Println("[CHALENGE]", r.Host, r.RemoteAddr, r.Method, maxURL(r.RequestURI))
			defer f.Close()
			io.Copy(w, f)
			return
		}

		// Search this host.
		if !s.hosts[r.Host] {
			log.Println("[UNKNOWN HOST]", r.Host, r.RemoteAddr, r.Method, maxURL(r.RequestURI))
			http.Error(w, "Host not found", http.StatusNotFound)
			return
		}

		// redirect to secure hosts
		log.Println("[UNSECURE]", r.Host, r.RemoteAddr, r.Method, maxURL(r.RequestURI))
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI,
			http.StatusPermanentRedirect)
	}))

	log.Fatal("[LISTEN ERROR]", a, err)
}

func maxURL(u string) string {
	if len(u) > 60 {
		return u[:60] + "..."
	}
	return u
}

func openChalenge(chalenge, p string) (http.File, bool) {
	f, e1 := http.Dir(chalenge).Open(p)
	if e1 != nil {
		return nil, false
	}

	if stats, err := f.Stat(); err != nil || stats.IsDir() {
		f.Close()
		return nil, false
	}

	return f, true
}
