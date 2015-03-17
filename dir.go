package fs

import (
	"os"
)

type Dir interface {
	Path() string
	Parent() Dir
	Up(level int) Dir
	Name() string
	Remove() error
	RemoveAll() error
	Exists() (bool, error)
	Rename(string) error
	Join(...string) string
	// like os.File.ReadDir
	Files() ([]File, error)
	Dirs() ([]Dir, error)
	// like filepath.Walk
	Create() error
	CreateAll() error
	Walk(func(File) error, func(Dir) error) error
}

type LocalDir interface {
	Move(dir string) error
	LoadInfo() error
	Info() (os.FileInfo, error)
	ReadDir(n int) ([]os.FileInfo, error)
}
