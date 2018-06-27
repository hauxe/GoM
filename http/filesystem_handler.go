package http

import (
	"net/http"
	"strings"
)

// GetFileSystemHandler get the file system http handler
func GetFileSystemHandler(directory, path string) http.HandlerFunc {
	fileServer := http.FileServer(FileSystem{http.Dir(directory)})
	return http.StripPrefix(strings.TrimRight(path, "/"), fileServer).ServeHTTP
}

// FileSystem custom file system handler
type FileSystem struct {
	fs http.FileSystem
}

// Open opens file
func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}
