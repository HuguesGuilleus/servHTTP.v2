// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	l, err := tls.Listen("tcp", a, &s.Config)
	if err != nil {
		log.Fatal("[LISTEN ERROR]", a, err)
		return
	}
	log.Println("[LISTEN TLS]", a)

	serv := http.Server{
		Handler: &s.m,
		ErrorLog: log.New(log.Writer(),
			"[LISTEN ERROR] <"+a+"> ",
			log.Flags()|log.Lmsgprefix),
	}
	serv.Serve(l)
}

// Serve to redirect all hosts to https
func (s *S) Serve(a, chalenge string) {
	log.Println("[LISTEN UNSAFE]", a, chalenge)
	err := http.ListenAndServe(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to response with a chalenge.
		if f := openChalenge(chalenge, r.URL.Path); f != nil {
			logReq("[CHALENGE]", r.Host, r)
			defer f.Close()
			io.Copy(w, f)
			return
		}

		// Search this host.
		if !s.hosts[r.Host] {
			logReq("[UNKNOWN HOST]", r.Host, r)
			http.Error(w, "Host not found", http.StatusNotFound)
			return
		}

		// redirect to secure hosts
		logReq("[2HTTPS]", r.Host, r)
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI,
			http.StatusPermanentRedirect)
	}))
	log.Fatal("[LISTEN ERROR]", a, err)
}

// Open and return a regular file for chalenge.
func openChalenge(chalenge, p string) http.File {
	f, e1 := http.Dir(chalenge).Open(p)
	if e1 != nil {
		return nil
	}

	if stats, err := f.Stat(); err != nil || stats.IsDir() {
		f.Close()
		return nil
	}

	return f
}

// Log a requet
func logReq(prefix, rhost string, r *http.Request) {
	u := r.RequestURI
	if len(u) > 100 {
		u = u[:100] + "..."
	}

	h := ""
	if rhost != "" {
		h = " (host:" + strconv.Quote(rhost) + ")"
	}

	log.Println(prefix, "=>"+h,
		strings.SplitN(r.RemoteAddr, ":", 2)[0],
		r.Method,
		u,
	)
}
