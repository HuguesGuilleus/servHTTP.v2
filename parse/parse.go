// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"errors"
	"fmt"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/handlers"
	"strings"
)

var (
	NeedArgs        = errors.New("Need args")
	NeedKey         = errors.New("Need a key")
	NeedDefaultHost = errors.New("Need a default host")
	NeedPath        = errors.New("Nedd path (expected host/path)")
	NotPlugin       = errors.New("Not a plugin declaration")
	NotAbsolute     = errors.New("Expected a absolute plugin path")
	IsPlugin        = errors.New("It's a plugin declaration.")
)

// Get the host and the path from "host/path" or from defaultHost + key= "/path"
func GetHostPath(defaultHost, key string) (host, path string, err error) {
	if key == "" {
		return "", "", NeedKey
	} else if i := strings.IndexByte(key, '/'); i == 0 {
		if defaultHost == "" {
			return "", "", NeedDefaultHost
		}
		return defaultHost, key, nil
	} else if i == -1 {
		return "", "", NeedPath
	} else {
		host = key[:i]
		path = key[i:]
		if path == "" {
			return "", "", NeedPath
		}
		return
	}
}

// Return the plugin and it's args.
func Plugin(inputs []string) (plugin string, args []string, err error) {
	if len(inputs) < 2 {
		return "", nil, NeedArgs
	} else if inputs[0] != "plugin" {
		return "", nil, NotPlugin
	}
	plugin = inputs[1]
	args = inputs[2:]
	if !strings.HasPrefix(inputs[1], "/") {
		return "", nil, NotAbsolute
	}
	return
}

func Args(args []string) (ht serv.HandlerType, target string, headers handlers.Headers, err error) {
	if len(args) == 0 {
		return 0, "", nil, NeedArgs
	} else if strings.HasPrefix(args[0], "/") {
		ht = serv.HandlerFile
		target = args[0]
		args = args[1:]
	} else {
		switch args[0] {
		case "file":
			ht = serv.HandlerFile
		case "redirect":
			ht = serv.HandlerRedirect
		case "reverse":
			ht = serv.HandlerReverse
		case "message":
			ht = serv.HandlerMessage
		default:
			return 0, "", nil, fmt.Errorf("Unknwo this handler: %q", args[0])
		}
		target = args[1]
		args = args[2:]
	}

	for i, h := range args {
		if err := parseHeaders(h, &headers); err != nil {
			return 0, "", nil, fmt.Errorf("args[%d]: %w", i+2, err)
		}
	}

	return
}

// Parse the header lines h, and add it to the list.
func parseHeaders(h string, list *handlers.Headers) error {
	for l, h := range strings.Split(h, "\n") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		split := strings.IndexByte(h, ':')
		if split == -1 {
			return fmt.Errorf("Expected a header line ('key: value') [%d] %q", l+1, h)
		}
		key := strings.TrimSpace(h[:split])
		value := strings.TrimSpace(h[split+1:])
		if key == "" || value == "" {
			return fmt.Errorf("Expected a header line ('key: value') [%d] %q", l+1, h)
		}
		*list = append(*list, struct{ Key, Value string }{key, value})
	}
	return nil
}
