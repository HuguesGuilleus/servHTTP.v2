// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/HuguesGuilleus/go-logoutput"
	"github.com/HuguesGuilleus/servHTTP.v2/parse"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg"
	"gopkg.in/ini.v1"
	"log"
	"strings"
)

func main() {
	configFile := flag.String("c", "/etc/servHTTP.ini", "The configuration file.")
	flag.Parse()

	log.SetFlags(log.Ltime)
	log.Println("[START]")
	config, err := ini.Load(*configFile)
	if err != nil {
		log.Fatal("[CONFIG ERROR]", err)
	}

	out := logoutput.New(config.Section("").Key("log").MustString("/var/log/servHTTP/"))
	log.SetOutput(out)
	server := serv.New(out, config.Section("").Key("noHttpLog").MustBool(false))

	for _, s := range config.Sections() {
		if n := s.Name(); n == "DEFAULT" {
			continue
		} else if strings.HasPrefix(n, "!cert") {
			if err := server.AddCert(s.Key("crt").String(), s.Key("key").String()); err != nil {
				log.Fatalf("[LOAD CERT/KEY ERROR] in section %q: %v\n", n, err)
			}
		} else {
			for _, k := range s.Keys() {
				line := k.Strings(" ")
				if err := loadLine(&server, n, k.Name(), line); err != nil {
					log.Fatalf("[LOAD ERROR] Insection %q, %q: %q: %v", n, k.Name(), line, err)
				}
			}
		}
	}

	if !config.Section("").Key("notls").MustBool() {
		go server.ListenTLS(config.Section("").Key("addrtls").MustString(":443"))
	}

	go server.ListenUnsecure(
		config.Section("").Key("addr").MustString(":80"),
		config.Section("").Key("chalenge").MustString("/var/letsencrypt"),
	)

	select {}
}

func loadLine(server *serv.Server, section, key string, line []string) error {
	host, path, err := parse.GetHostPath(section, key)
	if err != nil {
		return err
	}

	ht, target, headers, err := parse.Args(line)
	if err == nil {
		return server.AddHandler(ht, host, path, target, headers)
	}

	if err == parse.IsPlugin {
		plugin, args, err := parse.Plugin(line)
		if err != nil {
			return err
		}
		return server.AddPlugin(host, path, plugin, args)
	}

	return err
}
