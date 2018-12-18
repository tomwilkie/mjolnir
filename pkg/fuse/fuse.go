package fuse

import (
	"fmt"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/spf13/afero"
)

// Fuse Fuse Fuse.
type Fuse struct {
	conn *fuse.Conn
	fs   afero.Fs
}

// Config for Fuse.
type Config struct {
	MountPath string
}

// New Fuse.
func New(cfg Config, afs afero.Fs) {
	conn, err := fuse.Mount(cfg.MountPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	f := &Fuse{
		conn: conn,
		fs:   afs,
	}
	fs.Serve(f)
}
