// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Logger struct {
	writer io.Writer
	pool   sync.Pool
}

// Create a simple stdout logger.
func Default() Logger {
	return New(os.Stdout)
}

// Create a new logger to stdout.
func New(w io.Writer) Logger {
	return Logger{
		writer: w,
		pool: sync.Pool{
			New: func() interface{} {
				b := bytes.NewBuffer(nil)
				b.Grow(512)
				return b
			},
		},
	}
}

// Log one operation.
func (l *Logger) Operation(op string, args ...string) {
	b := l.begin(op)
	defer l.pool.Put(b)

	for _, a := range args {
		b.WriteByte(' ')
		b.WriteString(a)
	}
	b.WriteByte('\n')

	b.WriteTo(l.writer)
}

// Log on HTTP operation.
func (l *Logger) HTTP(op string, r *http.Request) {
	b := l.begin(op)
	defer l.pool.Put(b)

	path := r.URL.Path
	if len(path) > 80 {
		path = path[:80]
	}

	b.WriteByte(' ')
	b.WriteString(r.Method)
	b.WriteByte(' ')
	b.WriteString(r.URL.Host)
	b.WriteString(path)
	b.WriteByte('\n')

	b.WriteTo(l.writer)
}

// Print the begin of the line and return the buffer.
func (l *Logger) begin(op string) (b *bytes.Buffer) {
	b = l.pool.Get().(*bytes.Buffer)
	b.Reset()

	b.WriteString(time.Now().UTC().Format("2006-01-02 15:04:05 ["))
	b.WriteString(op)
	b.WriteByte(']')

	return
}
