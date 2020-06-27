// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"./pkg"
	"github.com/HuguesGuilleus/go-logoutput"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"path"
	"strings"
)

func init() {
	for _, a := range os.Args[1:] {
		if a == "-h" || a == "--help" || a == "--version" {
			log.SetFlags(0)
			log.Println("Give one config file in params or it's /etc/servHTTP.conf")
			os.Exit(0)
		}
	}
}

var Server serv.S

func main() {
	log.Println("[START]")

	// configFile := "conf.ini"
	configFile := "/etc/servHTTP.ini"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := ini.Load(configFile)
	if err != nil {
		log.Fatal("[CONFIG ERROR]", err)
	}

	logoutput.SetLog(config.Section("").Key("log").MustString("/var/log/servHTTP/"))

	for _, s := range config.Sections() {
		if n := s.Name(); n == "DEFAULT" {
			continue
		} else if strings.HasPrefix(n, "!cert") {
			err := Server.AddCert(s.Key("crt").String(), s.Key("key").String())
			if err != nil {
				log.Fatalf("[LOAD CERT/KEY ERROR] in secion %q: %v\n", n, err)
			}
		} else {
			for _, k := range s.Keys() {
				AddRule(n, k.Name(), k.Strings(" "))
			}
		}
	}

	if !config.Section("").Key("notls").MustBool() {
		go Server.ServeTLS(config.Section("").Key("addrtls").MustString(":443"))
	}
	go Server.Serve(
		config.Section("").Key("addr").MustString(":80"),
		config.Section("").Key("chalenge").MustString("/var/letsencrypt"),
	)
	select {}
}

// Add one rule.
func AddRule(h, p string, arg []string) {
	if !path.IsAbs(p) {
		log.Fatalf("[CONFIG ERORR] in host %q, path %q is not abolute\n", h, p)
	}

	var add func(string, string, string) error // A Server Add handler
	var a string                               // the argument to pass to add
	if len(arg) > 1 {
		a = arg[1]
		switch arg[0] {
		case "file":
			add = Server.AddFile
		case "reverse":
			add = Server.AddReverse
		case "redirect":
			add = Server.AddRedirect
		default:
			log.Fatalf("[CONFIG ERORR] in host %q path %q, unknwon handler %q\n", h, p, arg[0])
		}
	} else {
		add = Server.AddFile
		a = arg[0]
	}

	if err := add(h, p, a); err != nil {
		log.Fatalf("[CONFIG ERORR] in host %q path %q: %v\n", h, p, err)
	}
}
