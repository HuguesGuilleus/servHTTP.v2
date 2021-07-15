// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/handlers"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/logger"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("a", ":8000", `HTTP listen address`)
	key := flag.String("k", "", "TLS key(PEM) (or nothing to no encrypt)")
	cert := flag.String("c", "", "TLS certificate")
	flag.Parse()
	root := flag.Arg(0)
	if root == "" {
		root = "."
	}

	l := logger.Default()
	serv := &handlers.FileServer{
		Base: handlers.Base{
			Logger: &l,
			Base:   "/",
		},
		Root: root,
	}

	log.SetFlags(0)
	log.Println("\033[H\033[2J")
	l.Operation("listen", *addr)

	if *key != "" && *cert != "" {
		log.Fatal(http.ListenAndServeTLS(*addr, *cert, *key, serv))
	} else {
		log.Fatal(http.ListenAndServe(*addr, serv))
	}
}
