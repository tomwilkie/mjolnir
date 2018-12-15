package mjölnir

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (m *Mjölnir) Create(name string) (afero.File, error) {
	return nil, nil
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (m *Mjölnir) Mkdir(path string, perm os.FileMode) error {
	ctx := context.Background()
	path = filepath.Clean(path)
	dir, file := filepath.Split(path)

	return m.txn(ctx, func(t *txn) error {
		dirInode, err := t.lookup(dir)
		if err != nil {
			return err
		}
		if !dirInode.IsDir {
			return os.ErrInvalid
		}

		newInode := inode{
			id: rand.Uint64(),
			Inode: Inode{
				Perms: uint32(perm),
				IsDir: true,
			},
		}

		if err := dirInode.add(file, newInode.id); err != nil {
			return err
		}

		t.put(dirInode)
		t.put(&newInode) // todo make sure we don't overwire with ID collision.
		return nil
	})
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (m *Mjölnir) MkdirAll(path string, perm os.FileMode) error {
	path = filepath.Clean(path)
	base := ""
	for _, segment := range filepath.SplitList(path) {
		base = filepath.Join(base, segment)
		err := m.Mkdir(base, perm)
		if err == os.ErrExist {
			continue
		} else if err != nil {
			return err
		}
	}
	return nil
}

// Open opens a file, returning it or an error, if any happens.
func (m *Mjölnir) Open(path string) (afero.File, error) {
	path = filepath.Clean(path)
	inode, err := m.lookup(context.Background(), path)
	if err != nil {
		return nil, err
	}
	return &file{
		path:  path,
		inode: inode,
	}, nil
}

// OpenFile opens a file using the given flags and the given mode.
func (m *Mjölnir) OpenFile(path string, _ int, _ os.FileMode) (afero.File, error) {
	path = filepath.Clean(path)
	inode, err := m.lookup(context.Background(), path)
	if err != nil {
		return nil, err
	}
	return &file{
		path:  path,
		inode: inode,
	}, nil
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (m *Mjölnir) Remove(path string) error {
	ctx := context.Background()
	path = filepath.Clean(path)
	dir, file := filepath.Split(path)

	return m.txn(ctx, func(t *txn) error {
		inode, err := t.lookup(dir)
		if err != nil {
			return err
		}
		if !inode.IsDir {
			return os.ErrInvalid
		}

		inodeID, err := inode.remove(file)
		if err != nil {
			return err
		}

		t.put(inode)
		t.delete(inodeID)
		return nil
	})
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (m *Mjölnir) RemoveAll(path string) error {
	return nil
}

// Rename renames a file.
func (m *Mjölnir) Rename(oldname, newname string) error {
	ctx := context.Background()
	oldname, newname = filepath.Clean(oldname), filepath.Clean(newname)
	olddir, oldfile := filepath.Split(oldname)
	newdir, newfile := filepath.Split(newname)

	return m.txn(ctx, func(t *txn) error {
		oldDirInode, err := t.lookup(olddir)
		if err != nil {
			return err
		}
		if !oldDirInode.IsDir {
			return os.ErrInvalid
		}

		inodeID, err := oldDirInode.remove(oldfile)
		if err != nil {
			return err
		}

		newDirInode := oldDirInode
		if olddir != newdir {
			newDirInode, err := t.lookup(newdir)
			if err != nil {
				return err
			}
			if !newDirInode.IsDir {
				return os.ErrInvalid
			}
		}

		// Add entry to new inode.
		if err := newDirInode.add(newfile, inodeID); err != nil {
			return err
		}

		t.put(oldDirInode)
		t.put(newDirInode)
		return nil
	})
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (m *Mjölnir) Stat(path string) (os.FileInfo, error) {
	ctx := context.Background()
	path = filepath.Clean(path)
	inode, err := m.lookup(ctx, path)
	if err != nil {
		return nil, err
	}

	_, name := filepath.Split(path)

	return fileInfo{
		name:  name,
		mode:  os.FileMode(inode.Perms),
		size:  inode.Size,
		isDir: inode.IsDir,
	}, nil
}

// Name of this FileSystem
func (m *Mjölnir) Name() string {
	return "Mjölnir"
}

// Chmod changes the mode of the named file to mode.
func (m *Mjölnir) Chmod(path string, mode os.FileMode) error {
	ctx := context.Background()
	path = filepath.Clean(path)

	return m.txn(ctx, func(t *txn) error {
		inode, err := t.lookup(path)
		if err != nil {
			return err
		}

		inode.Perms = uint32(mode)
		t.put(inode)
		return nil
	})
}

// Chtimes changes the access and modification times of the named file
func (m *Mjölnir) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil
}
