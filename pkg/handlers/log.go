// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
)

func (b *Base) HTTP(op logOperation, r *http.Request) {
	b.Logger.HTTP(op.String(), r)
}

type logOperation int

const (
	// unsafe
	// secure

	logFileServed logOperation = iota
	logFileCompressed
	logFileErrorMethod
	logFileErrorNotFound
	logFileErrorInternal
	logFileIndex

	logPluginLoadOK
	logPluginLoadError
	logMessage
	logRedirect

	logReverseOk
	logReverseError
)

func (op logOperation) String() string {
	switch op {
	case logFileServed:
		return "file.serve"
	case logFileCompressed:
		return "file.compressed"
	case logFileErrorMethod:
		return "file.error.method"
	case logFileErrorNotFound:
		return "file.error.notfound"
	case logFileErrorInternal:
		return "file.error.internal"
	case logFileIndex:
		return "file.index"
	case logMessage:
		return "message"
	case logRedirect:
		return "redirect"
	case logReverseOk:
		return "reverse.ok"
	case logReverseError:
		return "reverse.error"
	default:
		return "unknownoperation"
	}
}
