package fs

import (
	"io"
	"os"
)

type File interface {
	Name() string
	Path() string
	Parent() Dir
	Remove() error
	Extension() string
	BareName() string
	Exists() (bool, error)
	Read(func(io.Reader) error) error
	// Write will create the file if it does not exist
	Write(func(io.Writer) error) error
	Rename(string) error
	MimeType() string
}

type LocalFile interface {
	// like os.OpenFile
	Copy(path string) error
	IO(flag int, perm os.FileMode) (*os.File, error)
	Move(dir string) error
	Info() (os.FileInfo, error)
	LoadInfo() error
}
