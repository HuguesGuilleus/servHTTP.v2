// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

//xgo:generate staticFile --dev -out.file=front/all.go -out.pkg=front Index=pkg/front/index.gohtml  IndexJs=pkg/front/index.html E404=pkg/front/e404.gohtml E502=pkg/front/e502.gohtml

//go:generate staticFile -out.file=front/all.go -out.pkg=front Index=front/index.gohtml IndexJs=front/index.html E404=front/e404.gohtml E502=front/e502.gohtml

import (
	"./front"
	"html/template"
	"log"
	"os"
	"sync"
	"time"
)

// The min duration between to check of static file
const StaticUpdate time.Duration = time.Second * 10

// All templates.
type Static struct {
	Index   func() *template.Template
	IndexJs func() []byte
	E404    func() []byte
	E502    func() []byte
}

// New Static with default value.
func NewStatic() *Static {
	return &Static{
		Index: func() *template.Template {
			return template.Must(
				template.New("").Parse(string(front.Index())),
			)
		},
		IndexJs: front.IndexJs,
		E404:    front.E404,
		E502:    front.E502,
	}
}

// Return a function never nil that return the content of the file p or
// the default content if error or p is empty.
func StaticLoad(p string, d []byte) func() []byte {
	if p == "" {
		return func() []byte { return d }
	}

	var lastRead time.Time
	var lastMod time.Time
	var mutex sync.Mutex
	var data []byte = d

	return func() []byte {
		mutex.Lock()
		defer mutex.Unlock()
		if lastRead.After(time.Now()) {
			return data
		}
		defer func() { lastRead = time.Now().Add(StaticUpdate) }()

		stat, err := os.Stat(p)
		if err != nil {
			log.Println("[STATIC ERROR]", p, err)
			data = d
			return data
		}
		if lastMod.Equal(stat.ModTime()) {
			return data
		}
		lastMod = stat.ModTime()

		f, err := os.Open(p)
		if err != nil {
			log.Println("[STATIC ERROR]", p, err)
			data = d
			return data
		}
		defer f.Close()
		data = make([]byte, stat.Size(), stat.Size())
		f.Read(data)

		return data
	}
}

// Return a function never nil that return the content of the file p or
// teh default content in error or if p is empty.
func StaticLoadTempl(p string, d *template.Template) func() *template.Template {
	if p == "" {
		return func() *template.Template { return d }
	}

	var lastRead time.Time
	var lastMod time.Time
	var mutex sync.Mutex
	var templ *template.Template = d

	return func() *template.Template {
		mutex.Lock()
		defer mutex.Unlock()
		if lastRead.After(time.Now()) {
			return templ
		}
		defer func() { lastRead = time.Now().Add(StaticUpdate) }()

		// Check if the file are modified
		stat, err := os.Stat(p)
		if err != nil {
			log.Println("[STATIC READ ERROR]", p, err)
			templ = d
			return templ
		}
		if lastMod.Equal(stat.ModTime()) {
			return templ
		}
		lastMod = stat.ModTime()

		// Update the template
		f, err := os.Open(p)
		if err != nil {
			log.Println("[STATIC READ ERROR]", p, err)
			templ = d
			return templ
		}
		defer f.Close()
		data := make([]byte, stat.Size(), stat.Size())
		f.Read(data)
		t, err := template.New("").Parse(string(data))
		if err != nil {
			log.Println("[STATIC TEMPLATE ERROR]", p, err)
			templ = d
			return templ
		}
		templ = t

		return templ
	}
}
