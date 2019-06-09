package posix

import (
	"path/filepath"
	"strings"
)

const separator = string(filepath.Separator) // OS-specific path separator

// pathClean is equivalent to but slightly more efficient than
// filepath.Clean(separator + path). If path is empty pathClean returns
// filepath.Separator.
func pathClean(path string) string {
	if path == "" || path[0] != filepath.Separator {
		path = separator + path
	}

	return filepath.Clean(path)
}

// clean is equivalent to filepath.Clean(name) but removes all leading
// filepath.Separator and hence returns a relative path.
func clean(name string) string {
	for len(name) > 0 && name[0] == filepath.Separator {
		name = name[1:]
	}

	return filepath.Clean(name)
}

// isAbs reports whether the path is absolute.
func isAbs(path string) bool { return filepath.IsAbs(path) }

// Join joins any number of path elements into a single path, adding a
// separator if necessary. Join calls Clean on the result; in particular,
// all empty strings are ignored.
func join(elem ...string) string { return filepath.Join(elem...) }

// split slices path into all names separated by filepath.Separator and
// returns a slice of the path elements between filepath.Separator.
// If path is empty or represents the root split returns an empty slice.
func split(path string) []string {
	for len(path) > 0 && path[0] == filepath.Separator {
		path = path[1:]
	}

	if len(path) == 0 || path == separator {
		return []string{}
	}
	return strings.Split(path, separator)
}

// hasPrefix will determine if path starts with prefix from the point of
// view of a filesystem.
func hasPrefix(path, root string) bool {
	if root == "" || root == separator {
		return true
	}

	names, prefix := split(path), split(root)
	if len(prefix) > len(names) {
		return false
	}

	for i, name := range prefix {
		if names[i] != name {
			return false
		}
	}
	return true
}

// isReserved returns whetever name is a reserved filesystem name.
func isReserved(name string) bool {
	return name == "" || name == "." || name == ".."
}

// chroot joins root and path into a single path, adding a separator if
// necessary. chroot returns false if the result does not start with
// root from the point of view of a filesystem.
func chroot(root, path string) (string, bool) {
	if root == "" || root == separator {
		return pathClean(path), true
	}

	path = filepath.Join(root, clean(path))
	if !hasPrefix(path, root) {
		return "", false
	}
	return path, true
}
