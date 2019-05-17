package fs

import (
	"reflect"
	"testing"
)

func TestPathClean(t *testing.T) {
	for num, test := range []struct {
		path string
		want string
	}{
		{"//a///b///c", "/a/b/c"},
		{"/a/b/c/.", "/a/b/c"},
		{"/a/b/c", "/a/b/c"},
		{"a/b/c", "/a/b/c"},

		{"/a/../b", "/b"},
		{"a/../b", "/b"},

		{"/", "/"},
		{"", "/"},
		{".", "/"},
		{"..", "/"},
		{"/..", "/"},
		{"/../.", "/"},
		{"/../..", "/"},
	} {
		path := pathClean(test.path)

		if path != test.want {
			t.Fatalf("pathClean(%d): expeceted path %q, got %q", num, test.want, path)
		}
	}
}

func TestClean(t *testing.T) {
	for num, test := range []struct {
		path string
		want string
	}{
		{"//a///b///c", "a/b/c"},
		{"/a/b/c/.", "a/b/c"},
		{"/a/b/c", "a/b/c"},
		{"a/b/c", "a/b/c"},

		{"/a/../b", "b"},
		{"a/../b", "b"},

		{"/", "."},
		{"", "."},

		{".", "."},
		{"..", ".."},
		{"/..", ".."},
		{"/../.", ".."},
		{"/../..", "../.."},
	} {
		path := clean(test.path)

		if path != test.want {
			t.Fatalf("clean(%d): expeceted path %q, got %q", num, test.want, path)
		}
	}
}

func TestPathSplit(t *testing.T) {
	for num, test := range []struct {
		path string
		want []string
	}{
		{"///a/b/c", []string{"a", "b", "c"}},
		{"/a/b/c", []string{"a", "b", "c"}},
		{"a/b/c", []string{"a", "b", "c"}},

		{"/a/b", []string{"a", "b"}},
		{"a/b", []string{"a", "b"}},

		{"/a", []string{"a"}},
		{"a", []string{"a"}},

		{"a///b", []string{"a", "", "", "b"}},
		{"a//b", []string{"a", "", "b"}},

		{"", []string{}},
	} {
		names := split(test.path)
		if !reflect.DeepEqual(names, test.want) {
			t.Errorf("split(%d): expected names %v, got %v", num, test.want, names)
		}
	}
}

func TestHasPrefix(t *testing.T) {
	for num, test := range []struct {
		path   string
		prefix string
		want   bool
	}{
		{"/a/b/c", "/a/b/c", true},
		{"/a/b/c/", "/a/b/c", true},
		{"a/b/c", "/a/b/c", true},

		{"/", "/", true},
		{"", "", true},

		{"/a/b/c", "/a/b", true},
		{"/a/ba", "/a/b", false},
		{"/a/ba/c", "/a/b", false},

		{"/a", "/a/b", false},
	} {
		res := hasPrefix(test.path, test.prefix)

		if res != test.want {
			t.Fatalf("hasPrefix(%d): differ %q, got %q", num, test.prefix, test.path)
		}
	}
}
