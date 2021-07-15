// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/handlers"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/logger"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"plugin"
)

type HandlerType int

const (
	HandlerFile HandlerType = iota
	HandlerRedirect
	HandlerReverse
	HandlerMessage
	HandlerPlugin
)

type Server struct {
	mux   http.ServeMux
	certs []tls.Certificate
	Log   *logger.Logger
	hosts map[string]bool
	// Output for main log and plugins.
	logOutput io.Writer
	// Disable http log error
	NoHTTPOutput bool
}

func New(w io.Writer, noHTTPOutput bool) Server {
	l := logger.New(w)
	return Server{
		Log:       &l,
		logOutput: w,
		hosts:     make(map[string]bool),
	}
}

// Add a new cert to the list.
func (s *Server) AddCert(key, crt string) error {
	cert, err := tls.LoadX509KeyPair(key, crt)
	if err != nil {
		return err
	}
	s.certs = append(s.certs, cert)

	return nil
}

func (s *Server) AddPlugin(host, path string, pluginPath string, args []string) error {
	plugin, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("Load plugin %q fail: %w", pluginPath, err)
	}

	if symbol, err := plugin.Lookup("Plug"); err != nil {
		return fmt.Errorf("Fail to load PLug symbol: %w", err)
	} else if plug, ok := symbol.(func(w io.Writer, args []string) (http.Handler, error)); !ok {
		return fmt.Errorf("Plug is not a symbol: %[1]T %[1]v", symbol)
	} else if handler, err := plug(s.logOutput, args); err != nil {
		return fmt.Errorf("Fail fo plug %q with %q: %w", pluginPath, args, err)
	} else {
		s.Handle(host, path, http.StripPrefix(path, handler))
	}
	return nil
}

func (s *Server) AddHandler(ht HandlerType, host, path string, target string, headers handlers.Headers) error {
	base := handlers.Base{
		Logger:  s.Log,
		Base:    path,
		Headers: headers,
	}

	switch ht {
	case HandlerFile:
		if !filepath.IsAbs(target) {
			return fmt.Errorf("The root is not absolute: %q", target)
		}
		s.Handle(host, path, &handlers.FileServer{
			Base: base,
			Root: target,
		})
	case HandlerRedirect:
		u, err := url.Parse(target)
		if err != nil {
			return err
		}
		s.Handle(host, path, &handlers.RedirectServer{
			Base: base,
			URL:  u,
		})
	case HandlerReverse:
		u, err := url.Parse(target)
		if err != nil {
			return err
		}
		s.Handle(host, path, handlers.Reverse(base, u))
	case HandlerMessage:
		s.Handle(host, path, &handlers.MessageServer{
			Base:    base,
			Message: target,
		})
	case HandlerPlugin:
		return errors.New("Can not load plugin with this methode, use AddPlugin insteed.")
	default:
		return fmt.Errorf("Unknow this hanlder: %d, for %s%s", ht, host, path)
	}
	return nil
}

// Add the handler for the host and path.
func (s *Server) Handle(host, path string, handler http.Handler) {
	s.hosts[host] = true
	s.mux.Handle(host+path, handler)
}

// ListenTLS and server HTTPS response with added handlers.
func (s *Server) ListenTLS(addr string) error {
	l, err := tls.Listen("tcp", addr, &tls.Config{
		NextProtos:   []string{"h2", "http/1.1"},
		Certificates: s.certs,
	})
	if err != nil {
		return err
	}

	serv := http.Server{
		Handler: &s.mux,
	}
	if s.NoHTTPOutput {
		serv.ErrorLog = log.New(io.Discard, "", 0)
	}
	return serv.Serve(l)
}

// Listen unsecure request, response with a file, for chalenging, or if
// the host is known with a redirect to https.  If dir is empty, no file
// can be served.
func (s *Server) ListenUnsecure(addr, dir string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dir != "" {
			if f := regularFile(filepath.Join(dir, filepath.FromSlash(r.URL.Path))); f != nil {
				defer f.Close()
				s.Log.HTTP("unsecure.challenge", r)
				io.Copy(w, f)
				return
			}
		}

		if !s.hosts[r.Host] {
			s.Log.HTTP("unsecure.unknown", r)
			http.NotFound(w, r)
			return
		}

		s.Log.HTTP("unsecure.redirect", r)
		r.URL.Scheme = "https"
		r.URL.Host = r.Host
		http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
	}))
}

// Open the regular file at path. If error or directory return nil.
func regularFile(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		f.Close()
		return nil
	}
	return f
}
