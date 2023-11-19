//go:build !go1.16
// +build !go1.16

package fontscan

import (
	"os"
	"path/filepath"
)

// recursively walk through the given directory, scanning font files and calling dst.consume
// for each valid file found.
func (dst *footprintScanner) scanDirectory(logger Logger, dir string, visited map[string]bool) error {
	walkFn := func(path string, d os.FileInfo, err error) error {
		if err != nil {
			logger.Printf("error walking font directory %q: %v", path, err)
			return filepath.SkipDir
		}

		if d.IsDir() { // keep going
			return nil
		}

		if visited[path] {
			return nil // skip the path
		}
		visited[path] = true

		// load the information, following potential symoblic links
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		// always ignore files which should never be font files
		if ignoreFontFile(info.Name()) {
			return nil
		}

		err = dst.consume(path, info)

		return err
	}

	err := filepath.Walk(dir, walkFn)

	return err
}

// DirEntry is a copy of the Go 1.16+ fs.DirEntry interface.
type dirEntry interface {
	// Name returns the name of the file (or subdirectory) described by the entry.
	// This name is only the final element of the path (the base name), not the entire path.
	// For example, Name would return "hello.go" not "home/gopher/hello.go".
	Name() string

	// IsDir reports whether the entry describes a directory.
	IsDir() bool

	// Type returns the type bits for the entry.
	// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
	Type() os.FileMode

	// Info returns the FileInfo for the file or subdirectory described by the entry.
	// The returned FileInfo may be from the time of the original directory read
	// or from the time of the call to Info. If the file has been removed or renamed
	// since the directory read, Info may return an error satisfying errors.Is(err, ErrNotExist).
	// If the entry denotes a symbolic link, Info reports the information about the link itself,
	// not the link's target.
	Info() (os.FileInfo, error)
}

// dirEntryAdapter wraps a normal os.FileInfo to be compatible with the DirEntry interface.
type dirEntryAdapter struct {
	os.FileInfo
}

func (e dirEntryAdapter) Info() (os.FileInfo, error) {
	return e.FileInfo, nil
}

// Type copies the constants from modern Go's fs package and uses them to provide a file's
// Type.
// https://cs.opensource.google/go/go/+/refs/tags/go1.20.5:src/io/fs/fs.go;l=239
func (e dirEntryAdapter) Type() os.FileMode {
	// The defined file mode bits are the most significant bits of the FileMode.
	// The nine least-significant bits are the standard Unix rwxrwxrwx permissions.
	// The values of these bits should be considered part of the public API and
	// may be used in wire protocols or disk representations: they must not be
	// changed, although new bits might be added.
	const (
		// The single letters are the abbreviations
		// used by the String method's formatting.
		ModeDir        os.FileMode = 1 << (32 - 1 - iota) // d: is a directory
		ModeAppend                                        // a: append-only
		ModeExclusive                                     // l: exclusive use
		ModeTemporary                                     // T: temporary file; Plan 9 only
		ModeSymlink                                       // L: symbolic link
		ModeDevice                                        // D: device file
		ModeNamedPipe                                     // p: named pipe (FIFO)
		ModeSocket                                        // S: Unix domain socket
		ModeSetuid                                        // u: setuid
		ModeSetgid                                        // g: setgid
		ModeCharDevice                                    // c: Unix character device, when ModeDevice is set
		ModeSticky                                        // t: sticky
		ModeIrregular                                     // ?: non-regular file; nothing else is known about this file

		// Mask for the type bits. For regular files, none will be set.
		ModeType = ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice | ModeCharDevice | ModeIrregular
	)

	return e.FileInfo.Mode() & ModeType
}

// readDir re-implements os.ReadDir (Go 1.16+) using only Go 1.14's stdlib.
func readDir(name string) ([]dirEntry, error) {
	d, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	entries, err := d.Readdir(0)
	if err != nil {
		return nil, err
	}
	adapted := make([]dirEntry, len(entries))
	for i, e := range entries {
		adapted[i] = dirEntryAdapter{e}
	}
	return adapted, nil
}
