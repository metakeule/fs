package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type dir struct {
	name     string
	dir      string
	fileinfo os.FileInfo
}

func NewDir(path string) *dir {
	path = filepath.Clean(path)
	return &dir{
		name: filepath.Base(path),
		dir:  filepath.Dir(path),
	}
}

func (d *dir) Name() string {
	return d.name
}

func (d *dir) Path() string {
	return filepath.Join(d.dir, d.name)
}

func (d *dir) Parent() Dir {
	return NewDir(d.dir)
}

func (d *dir) Up(level int) Dir {
	return NewDir(filepath.Join(d.Path(), strings.Repeat("../", level)))
}

func (d *dir) Join(segm ...string) string {
	str := []string{d.Path()}
	str = append(str, segm...)
	return filepath.Join(str...)
}

// like os.File.ReadDir
func (d *dir) ReadDir(n int) ([]os.FileInfo, error) {
	dd, err := os.Open(d.Path())
	if err != nil {
		return nil, err
	}
	defer dd.Close()
	return dd.Readdir(n)
}

func (d *dir) Files() (files []File, err error) {
	var fi []os.FileInfo
	fi, err = d.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	for _, ff := range fi {
		if !ff.IsDir() {
			files = append(files, NewFile(filepath.Join(d.Path(), ff.Name())))
		}
	}
	return
}

func (d *dir) Dirs() (dirs []Dir, err error) {
	var fi []os.FileInfo
	fi, err = d.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	for _, dd := range fi {
		if dd.IsDir() {
			dirs = append(dirs, NewDir(filepath.Join(d.Path(), dd.Name())))
		}
	}
	return
}

// Create makes local dir with 0755 permissions
func (d *dir) Create() error {
	return os.Mkdir(d.Path(), os.FileMode(0755))
}

// CreateAll makes local dir and all intermediate missing dirs with 0755 permissions
func (d *dir) CreateAll() error {
	return os.MkdirAll(d.Path(), os.FileMode(0755))
}

func (d *dir) Remove() error {
	return os.Remove(d.Path())
}

func (d *dir) RemoveAll() error {
	return os.RemoveAll(d.Path())
}

var ErrIsFile = errors.New("is a file")

func (d *dir) Exists() (bool, error) {
	err := d.LoadInfo()
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	if err == ErrIsFile {
		return true, err
	}

	return false, err
}

func (d *dir) LoadInfo() (err error) {
	d.fileinfo, err = os.Stat(d.Path())
	if err != nil {
		return err
	}
	if !d.fileinfo.IsDir() {
		d.fileinfo = nil
		return ErrIsFile
	}
	return nil
}

func (d *dir) Info() (os.FileInfo, error) {
	if d.fileinfo == nil {
		err := d.LoadInfo()
		if err != nil {
			return nil, err
		}
	}
	return d.fileinfo, nil
}

/*
func (d *dir) hasInfo() bool {
	return d.fileinfo != nil
}
*/

func (d *dir) Rename(name string) error {
	mustName(name)
	oldname := d.name
	oldpath := d.Path()
	d.name = name
	err := os.Rename(oldpath, d.Path())
	if err != nil {
		d.name = oldname
	}
	return err
}

func (d *dir) Move(dir string) error {
	oldDir := d.dir
	oldpath := d.Path()
	d.dir = dir
	err := os.Rename(oldpath, d.Path())
	if err != nil {
		d.dir = oldDir
	}
	return err
}

// like filepath.Walk
func (d *dir) Walk(filefn func(File) error, dirfn func(Dir) error) error {
	return filepath.Walk(d.Path(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if dirfn != nil {
				d := NewDir(filepath.Join(d.Path(), path))
				d.fileinfo = info
				return dirfn(d)
			}
		}

		if filefn != nil {
			f := NewFile(filepath.Join(d.Path(), path))
			f.fileinfo = info
			return filefn(f)
		}
		return nil
	})
}

var _ LocalDir = &dir{}
