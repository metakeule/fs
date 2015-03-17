package fs

import (
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type file struct {
	name     string
	dir      string
	fileinfo os.FileInfo
}

func mustName(name string) {
	if strings.ContainsRune(name, os.PathSeparator) {
		panic("invalid name: " + name)
	}
}

func NewFile(path string) *file {
	path = filepath.Clean(path)
	return &file{
		name: filepath.Base(path),
		dir:  filepath.Dir(path),
	}
}

func (f *file) MimeType() string {
	return mime.TypeByExtension("." + f.Extension())
}

func (f *file) BareName() string {
	idx := strings.LastIndex(f.name, ".")
	if idx == -1 {
		return f.name
	}
	return f.name[:idx]
}

func (f *file) Extension() string {
	idx := strings.LastIndex(f.name, ".")
	if idx == -1 {
		return ""
	}
	return f.name[idx:]
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Path() string {
	return filepath.Join(f.dir, f.name)
}

var ErrIsDir = errors.New("is a directory")
var ErrIsNotRegular = errors.New("is not a regular file")

func (f *file) LoadInfo() (err error) {
	f.fileinfo, err = os.Stat(f.Path())
	if err != nil {
		return err
	}
	if f.fileinfo.IsDir() {
		f.fileinfo = nil
		return ErrIsDir
	}

	if !f.fileinfo.Mode().IsRegular() {
		f.fileinfo = nil
		return ErrIsNotRegular
	}
	return nil
}

func (f *file) Info() (os.FileInfo, error) {
	if f.fileinfo == nil {
		err := f.LoadInfo()
		if err != nil {
			return nil, err
		}
	}
	return f.fileinfo, nil
}

func (f *file) Parent() Dir {
	return NewDir(f.dir)
}

func (f *file) Remove() error {
	return os.Remove(f.Path())
}

// similar to http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
func (f *file) Copy(path string) error {
	return f.Read(func(r io.Reader) (err error) {
		var out *os.File
		out, err = os.Create(path)
		if err != nil {
			return
		}
		defer func() {
			cerr := out.Close()
			if err == nil {
				err = cerr
			}
		}()
		if _, err = io.Copy(out, r); err != nil {
			return
		}
		err = out.Sync()
		return
	})
}

/*
func (f *file) HasInfo() bool {
	return f.fileinfo != nil
}
*/

func (f *file) Exists() (bool, error) {
	err := f.LoadInfo()
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	if err == ErrIsDir {
		return true, err
	}

	return false, err
}

func (f *file) Read(fn func(io.Reader) error) error {
	fl, err := os.Open(f.Path())
	if err != nil {
		return err
	}
	defer fl.Close()
	return fn(fl)
}

// Write will create the file if it does not exist
func (f *file) Write(fn func(io.Writer) error) error {
	fl, err := os.Create(f.Path())
	if err != nil {
		return err
	}
	defer fl.Close()
	defer fl.Sync()
	return fn(fl)
}

func (f *file) Rename(name string) error {
	mustName(name)
	oldname := f.name
	oldpath := f.Path()
	f.name = name
	err := os.Rename(oldpath, f.Path())
	if err != nil {
		f.name = oldname
	}
	return err
}

func (f *file) Move(dir string) error {
	oldDir := f.dir
	oldpath := f.Path()
	f.dir = dir
	err := os.Rename(oldpath, f.Path())
	if err != nil {
		f.dir = oldDir
	}
	return err
}

func (f *file) IO(flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(f.Path(), flag, perm)
}

var _ LocalFile = &file{}
