package posix

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

func TestChroot(t *testing.T) {
	for num, test := range []struct {
		root, path string
		result     string
		want       bool
	}{
		{"/", "/a", "/a", true},
		{"/", "a", "/a", true},
		{"", "/a", "/a", true},
		{"", "a", "/a", true},

		{"/a", "/b", "/a/b", true},
		{"/a", "b", "/a/b", true},

		{"/a/b", "../b/c", "/a/b/c", true},

		{"/a/b", "/..", "", false},
		{"/a/b", "..", "", false},

		{"/a", "/..", "", false},
		{"/a", "..", "", false},
	} {
		path, success := chroot(test.root, test.path)
		if success != test.want {
			t.Errorf("chroot(%d): expected result %v, got %v", num, test.want, success)
		}
		if path != test.result {
			t.Errorf("chroot(%d): expected result %q, got %q", num, test.result, path)
		}
	}
}

func TestIsValidName(t *testing.T) {
	check := func(t *testing.T, result, want bool) {
		t.Helper()
		if result != want {
			t.Errorf("isValidName: expected result %v, got %v", want, result)
		}
	}

	check(t, isValidName(".."), false)
	check(t, isValidName("."), false)
	check(t, isValidName(""), false)

	check(t, isValidName("/"), false)
	check(t, isValidName("a/"), false)
	check(t, isValidName("a/b"), false)
	check(t, isValidName("/a/b"), false)
}
