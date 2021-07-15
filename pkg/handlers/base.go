// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/logger"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	ServerName        = "servHTTP/3"
	internalReadError = "Internal read error."
)

// The base of all implementation of server.
type Base struct {
	*logger.Logger
	Base string
	Headers
}

// Headers contain custom headers add to incoming request.
type Headers []struct{ Key, Value string }

// Write headers to the response
func (h Headers) writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Server", ServerName)
	for _, h := range h {
		w.Header().Add(h.Key, h.Value)
	}
}

// Remove the base of the path.
func (b *Base) Trim(path string) string {
	return strings.TrimPrefix(path, b.Base)
}

/* Message server */

// Response with a constante message
type MessageServer struct {
	Base
	Message string
}

func (m *MessageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.writeHeaders(w)
	m.HTTP(logMessage, r)
	template.ErrorString(w, r, http.StatusOK, m.Message)
}

/* Redirect server */

type RedirectServer struct {
	Base
	URL *url.URL
}

func (rs *RedirectServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rs.writeHeaders(w)
	rs.HTTP(logRedirect, r)

	r.URL.Path = rs.Trim(r.URL.Path)
	http.Redirect(w, r, rs.URL.ResolveReference(r.URL).String(), http.StatusPermanentRedirect)
}

func Reverse(b Base, u *url.URL) http.Handler {
	r := httputil.NewSingleHostReverseProxy(u)

	director := r.Director
	r.Director = func(req *http.Request) {
		director(req)
		req.URL.Path = b.Trim(req.URL.Path)
		if !strings.HasPrefix(req.URL.Path, "/") {
			req.URL.Path = "/" + req.URL.Path
		}
		req.URL.RawPath = ""
	}

	r.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Add("Server", ServerName)
		for _, h := range b.Headers {
			resp.Header.Add(h.Key, h.Value)
		}
		b.HTTP(logReverseOk, resp.Request)
		return nil
	}

	r.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		b.writeHeaders(w)
		b.HTTP(logReverseError, r)
		template.ErrorString(w, r, http.StatusBadGateway, err.Error())
	}

	return r
}
