package mjölnir

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

type file struct {
	path  string
	inode *inode
	afero.File
}

func (f *file) Name() string {
	return f.path
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *file) Readdirnames(n int) ([]string, error) {
	names := make([]string, 0, len(f.inode.Dirents))
	for _, ent := range f.inode.Dirents {
		names = append(names, ent.Name)
	}
	return names, nil
}

func (f *file) Stat() (os.FileInfo, error) {
	return fileInfo{
		name:  f.path,
		mode:  os.FileMode(f.inode.Perms),
		size:  f.inode.Size,
		isDir: f.inode.IsDir,
	}, nil
}

type fileInfo struct {
	name    string      // base name of the file
	size    int64       // length in bytes for regular files; system-dependent for others
	mode    os.FileMode // file mode bits
	modTime time.Time   // modification time
	isDir   bool        // abbreviation for Mode().IsDir()
}

func (f fileInfo) Name() string {
	return f.name
}
func (f fileInfo) Size() int64 {
	return f.size
}
func (f fileInfo) Mode() os.FileMode {
	return f.mode
}
func (f fileInfo) ModTime() time.Time {
	return f.modTime
}
func (f fileInfo) IsDir() bool {
	return f.isDir
}
func (f fileInfo) Sys() interface{} {
	return nil
}
