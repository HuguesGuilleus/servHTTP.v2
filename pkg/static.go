// Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package serv

import (
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
	Index func() *template.Template
	E404  func() []byte
	E502  func() []byte
}

// New Static with default value.
func NewStatic() *Static {
	return &Static{
		Index: staticTempl("", template.Must(
			template.New("").Parse(defaultIndex))),
		E404: static("", []byte(defaultE404)),
		E502: static("", []byte(defaultE502)),
	}
}

// Default template to index directory
const defaultIndex = `<!doctype html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1"><style>body{max-width:70em;margin:auto;padding:1em;font-family:monospace;font-size:xx-large;background:#d3d3d3}h1{margin:0}#link{display:table;padding:.2em .5em;background:#fff}a{color:#1e90ff;background:inherit;text-decoration:none}a:hover{color:#00008b;text-decoration:underline}#list{list-style:none;padding:0}.info{font-size:inherit;color:#0000004f}</style><title>Index</title></head><body><h1>Index</h1><div id=link></div><ul id=list>{{range .}}<li><span class=info>[{{.ModTime.UTC.Format "2006-01-02 15:04:05 UTC"}}]</span> {{if .IsDir }}<a href={{.Name}}/>{{.Name}}/</a>{{else}}<a href={{.Name}} download=download>{{.Name}}</a> <span class="info size">{{.Size}}</span>{{end}}</li>{{end}}</ul><script>document.addEventListener('DOMContentLoaded',()=>{document.getElementById('link').innerHTML=document.location.pathname.split('/').filter((v,i)=>!i||v).map((v,i,a)=>'<a href="'+a.slice(0,i+1).join('/')+'/">'+v+'/</a>').join('');document.title=document.location.pathname.replace(/\/$/,'').split('/').pop();document.querySelectorAll('.size').forEach(s=>{s.title=s.innerText+' o';let h=((n)=>{if(n<1000){return n+'\u00A0o';}else if(n<1000_000){return(n/1000).toFixed(1)+'\u00A0K';}else if(n<1000_000_000){return(n/1000_000).toFixed(1)+'\u00A0M';}else if(n<1000_000_000_000){return(n/1000_000_000).toFixed(1)+'\u00A0G';}else if(n<1000_000_000_000_000){return(n/1000_000_000_000).toFixed(1)+'\u00A0T';}})(new Number(s.innerText));s.innerText='('+h+')'});},{once:true});</script></body></html>`

// Default page for error 404
const defaultE404 = `<!doctype html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1"><style>body{font-family:monospace;font-size:xxx-large;text-align:center;background:#d3d3d3}h1{margin-top:30vh}#link{margin:auto;padding:.2em 1.5em;display:table;background:#fff}a{color:#1e90ff;background:inherit;text-decoration:none}a:hover{color:#00008b;text-decoration:underline}</style><title>Error 404</title></head><body><h1>Error 404: Not found</h1><div id=link></div><script>document.addEventListener('DOMContentLoaded',()=>{document.getElementById('link').innerHTML=document.location.pathname.split('/').filter((v,i)=>!i||v).map((v,i,a)=>'<a href="'+a.slice(0,i+1).join('/')+'/">'+v+'/</a>').join('')},{once:true,});</script></body></html>`

// Default page for error 502
const defaultE502 = `<!doctype html><html lang=en><head><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1"><style>body{font-family:monospace;font-size:xxx-large;text-align:center;background:#d3d3d3}h1{margin-top:30vh}#link{margin:auto;padding:.2em 1.5em;display:table;background:#fff}a{color:#1e90ff;background:inherit;text-decoration:none}a:hover{color:#00008b;text-decoration:underline}</style><title>Error 502</title></head><body><h1>Error 502: Bad gateway</h1><div id=link></div><script>document.addEventListener('DOMContentLoaded',()=>{document.getElementById('link').innerHTML=document.location.pathname.split('/').filter((v,i)=>!i||v).map((v,i,a)=>'<a href="'+a.slice(0,i+1).join('/')+'/">'+v+'/</a>').join('')},{once:true,});</script></body></html>`

// Return a function never nil that return the content of the file p or
// the default content if error or p is empty.
func static(p string, d []byte) func() []byte {
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
func staticTempl(p string, d *template.Template) func() *template.Template {
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
