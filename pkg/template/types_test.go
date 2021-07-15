package template

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURLFill(t *testing.T) {
	var u URL
	u.Fill("example.com", "/dir/subdir/file.txt")
	assert.Equal(t, URL{
		Host: "example.com",
		Path: "/dir/subdir/file.txt",
		Parents: []URLDirectory{
			{"", "/"},
			{"dir", "/dir/"},
			{"subdir", "/dir/subdir/"},
		},
		File: "file.txt",
	}, u)
}

func TestHumainSize(t *testing.T) {
	assert.Equal(t, "123 o", humainSize(123))
	assert.Equal(t, "123.4 Ko", humainSize(123_400))
	assert.Equal(t, "123.4 Mo", humainSize(123_400_000))
	assert.Equal(t, "123.4 Go", humainSize(123_400_000_000))
	assert.Equal(t, "123.4 To", humainSize(123_400_000_000_000))
}
