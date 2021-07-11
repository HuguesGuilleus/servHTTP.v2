// Copyright (c) 2021, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"github.com/HuguesGuilleus/servHTTP.v2/pkg"
	"github.com/HuguesGuilleus/servHTTP.v2/pkg/handlers"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetHostPath(t *testing.T) {
	test := func(defaultHost, key, host, path string) {
		h, p, err := GetHostPath(defaultHost, key)
		assert.NoError(t, err)
		assert.Equal(t, host, h)
		assert.Equal(t, path, p)
	}
	test("host", "/dir/subdir", "host", "/dir/subdir")
	test("", "host/dir/subdir", "host", "/dir/subdir")
	test("other", "host/dir/subdir", "host", "/dir/subdir")

	failure := func(defaultHost, key string) {
		_, _, err := GetHostPath(defaultHost, key)
		if err == nil {
			t.Errorf("Expected error with %q %q", defaultHost, key)
		}
	}
	failure("", "")
	failure("", "/dir/subdir")
	failure("", "host")
	failure("host", "")
}

func TestPlgin(t *testing.T) {
	p, args, err := Plugin([]string{"plugin", "/dir/subdir/file.so", "a1", "a2"})
	assert.NoError(t, err)
	assert.Equal(t, "/dir/subdir/file.so", p)
	assert.Equal(t, []string{"a1", "a2"}, args)

	p, args, err = Plugin([]string{"plugin", "/dir/subdir/file.so"})
	assert.NoError(t, err)
	assert.Equal(t, "/dir/subdir/file.so", p)
	assert.Equal(t, []string{}, args)

	failure := func(expected error, args ...string) {
		_, _, err := Plugin(args)
		assert.EqualError(t, expected, err.Error(), args)
	}
	failure(NeedArgs, "/fdgfh")
	failure(NotPlugin, "file", "/fdgfh")
	failure(NotAbsolute, "plugin", "file.so")
}

func TestArgs(t *testing.T) {
	test := func(htExpected serv.HandlerType, args ...string) {
		ht, target, headers, err := Args(append(args, "k1: v1", "k2: v2"))
		assert.NoError(t, err, args)
		assert.Equal(t, htExpected, ht)
		assert.Equal(t, "/target", target)
		assert.Equal(t, handlers.Headers{{"k1", "v1"}, {"k2", "v2"}}, headers)
	}
	test(serv.HandlerFile, "/target")
	test(serv.HandlerFile, "file", "/target")
	test(serv.HandlerRedirect, "redirect", "/target")
	test(serv.HandlerReverse, "reverse", "/target")
	test(serv.HandlerMessage, "message", "/target")
}

func TestParseHeadears(t *testing.T) {
	var list handlers.Headers
	err := parseHeaders(`key: value
	k2 : yolo

	`, &list)
	assert.NoError(t, err)
}
