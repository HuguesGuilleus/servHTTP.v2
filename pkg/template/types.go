// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"encoding/base64"
	"encoding/json"
	"path"
	"strconv"
	"strings"
	"time"
)

// Information to display in this page
type page struct {
	ErrStatus string
	// The error code, 0 if index
	ErrCode int
	// The title of the page
	// Title string
	// Information about the URL
	URL
	// The list of the files if it's a index page.
	Files []FileInfo
	// The file name of the readme if exist.
	Readme string
	// The page contain media element.
	HasMedia bool
}

type URL struct {
	// The host name and port if it non standard.
	Host string
	// The complete path (using / separator)
	Path string
	// Parents is path splited, you can use for display URL with link.
	//
	//	Example for: /dir/subdir/file
	//
	//	| Name   | Absolute     |
	//	| :----- | :----------- |
	//	|        | /            |
	//	| dir    | /dir/        |
	//	| subdir | /dir/subdir/ |
	Parents []URLDirectory
	// The file, can be empty if the URL end with a '/'.
	File string
}

// A parent directory in an URL Path.
type URLDirectory struct {
	// The name of this directory
	Name string
	// The absolute path of this directory
	Absolute string
}

// A simplification of fs.FileInfo.
type FileInfo struct {
	Name    string
	Size    int64
	SizeHum string
	IsDir   bool
	ModTime time.Time
	// Information is the file is a media. Use it to create player in JS.
	Media *Media
}

// All data information.
type Media struct {
	Audio string `json:"a,omitempty"`
	Image string `json:"i,omitempty"`
	Video string `json:"v,omitempty"`
}

// Encodde in JSON with error ignoring.
func (m *Media) JSON() string {
	j, _ := json.Marshal(m)
	return base64.StdEncoding.EncodeToString(j)
}

// setURL fills the url from a host and a path.
func (u *URL) setURL(h, p string) {
	u.Host = h
	u.Path = p

	dir := ""
	dir, u.File = path.Split(p)
	dir = strings.TrimSuffix(dir, "/")

	dirs := strings.Split(dir, "/")
	u.Parents = make([]URLDirectory, len(dirs))
	l := 0
	for i, p := range dirs {
		l += len(p) + 1
		u.Parents[i] = URLDirectory{
			Name:     p,
			Absolute: u.Path[:l],
		}
	}
}

// Return size for humain.
func humainSize(size int64) string {
	var power int64
	var ext string

	if size < 1_000 {
		return strconv.Itoa(int(size)) + " o"
	} else if size < 1_000_000 {
		ext = " Ko"
		power = 1_000
	} else if size < 1_000_000_000 {
		ext = " Mo"
		power = 1_000_000
	} else if size < 1_000_000_000_000 {
		ext = " Go"
		power = 1_000_000_000
	} else {
		ext = " To"
		power = 1_000_000_000_000
	}

	var buff strings.Builder
	buff.Grow(8)
	buff.WriteString(strconv.Itoa(int(size / power)))
	buff.WriteByte('.')
	buff.WriteByte('0' + byte(size*10/power%10))
	buff.WriteString(ext)
	return buff.String()
}
