package mjölnir

import (
	"os"
	"sort"
	"strconv"
	"time"

	proto "github.com/golang/protobuf/proto"
)

type inode struct {
	id      uint64
	version int64
	Inode
}

func (i inode) ID() string {
	return strconv.FormatUint(i.id, 16)
}

func (i inode) Marshal() (string, error) {
	buf, err := proto.Marshal(&i.Inode)
	return string(buf), err
}

func (i *inode) add(name string, id uint64) error {
	j := sort.Search(len(i.Dirents), func(j int) bool {
		return i.Dirents[j].Name < name
	})
	if j < len(i.Dirents) && i.Dirents[j].Name == name {
		return os.ErrExist
	}
	i.Dirents = append(i.Dirents, nil)
	copy(i.Dirents[j+1:], i.Dirents[j:])
	i.Dirents[j] = &DirEnt{
		Name: name,
		Id:   id,
	}
	return nil
}

func (i *inode) remove(name string) (uint64, error) {
	j := sort.Search(len(i.Dirents), func(j int) bool {
		return i.Dirents[j].Name < name
	})
	if j >= len(i.Dirents) || i.Dirents[j].Name != name {
		return 0, os.ErrNotExist
	}
	id := i.Dirents[j].Id
	i.Dirents = i.Dirents[:j+copy(i.Dirents[j:], i.Dirents[j+1:])]
	return id, nil
}

func (i *inode) stat(name string) os.FileInfo {
	return fileInfo{
		name:    name,
		mode:    os.FileMode(i.Mode),
		size:    i.Size,
		isDir:   i.IsDir,
		modTime: time.Unix(i.Mtime, 0),
	}
}
