package bfs

import (
	"io"
	"io/fs"
	"path/filepath"
	"time"
)

func newFile(target string) *File {
	return &File{
		path:  target,
		watch: map[string]Event{},
	}
}

type File struct {
	path    string
	data    []byte
	mode    fs.FileMode
	modTime time.Time
	watch   map[string]Event
}

func (f *File) Path() string {
	return f.path
}

func (f *File) Write(data []byte) {
	f.data = append(f.data, data...)
}

func (f *File) Mode(mode fs.FileMode) {
	f.mode = mode
}

func (f *File) Watch(pattern string, event Event) {
	f.watch[pattern] |= event
}

func (f *File) open(fsys FS, key, relative, path string) (fs.File, error) {
	// fsys.watch(path, f.Watch)
	return &openFile{path, f.data, f.mode, f.modTime, 0}, nil
}

type openFile struct {
	path    string
	data    []byte
	mode    fs.FileMode
	modTime time.Time
	offset  int64
}

func (f *openFile) Close() error {
	return nil
}

func (f *openFile) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}
	if f.offset < 0 {
		return 0, &fs.PathError{Op: "read", Path: f.path, Err: fs.ErrInvalid}
	}
	n := copy(b, f.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *openFile) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name:    filepath.Base(f.path),
		data:    f.data,
		mode:    f.mode,
		modTime: f.modTime,
	}, nil
}

func (f *openFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		// offset += 0
	case 1:
		offset += f.offset
	case 2:
		offset += int64(len(f.data))
	}
	if offset < 0 || offset > int64(len(f.data)) {
		return 0, &fs.PathError{Op: "seek", Path: f.path, Err: fs.ErrInvalid}
	}
	f.offset = offset
	return offset, nil
}

type GenerateFile func(f FS, file *File) error

func (fn GenerateFile) open(f FS, key, relative, target string) (fs.File, error) {
	file := newFile(target)
	if err := fn(f, file); err != nil {
		return nil, err
	}
	for to, event := range file.watch {
		f.link(file.path, to, event)
	}
	return file.open(f, key, relative, target)
}

type fileGenerator interface {
	GenerateFile(f FS, file *File) error
}

func FileGenerator(generator fileGenerator) Generator {
	return GenerateFile(generator.GenerateFile)
}

type ServeFile func(f FS, file *File) error

func (fn ServeFile) open(f FS, key, relative, target string) (fs.File, error) {
	if relative == "." {
		return nil, fs.ErrInvalid
	}
	file := newFile(target)
	if err := fn(f, file); err != nil {
		return nil, err
	}
	for to, event := range file.watch {
		f.link(file.path, to, event)
	}
	return file.open(f, key, relative, target)
}

type fileServer interface {
	ServeFile(f FS, file *File) error
}

func FileServer(server fileServer) Generator {
	return GenerateFile(server.ServeFile)
}