package mjölnir

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/clientv3"
)

const rootInodeID = 0

// Mjölnir Mjölnir Mjölnir.
type Mjölnir struct {
	etcdClient *clientv3.Client
	kv         clientv3.KV
}

// New Mjölnir.
func New() (*Mjölnir, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		return nil, err
	}

	return &Mjölnir{
		etcdClient: client,
		kv:         clientv3.NewKV(client),
	}, nil
}

func (m *Mjölnir) get(ctx context.Context, id uint64) (*inode, error) {
	resp, err := m.kv.Get(ctx, strconv.FormatUint(id, 16))
	if err != nil {
		return nil, err
	}
	i := inode{
		id:      id,
		version: resp.Kvs[0].Version,
	}
	if err := proto.Unmarshal([]byte(resp.Kvs[0].Value), &i.Inode); err != nil {
		return nil, err
	}
	return &i, nil
}

func (m *Mjölnir) lookup(ctx context.Context, path string) (*inode, error) {
	inode, err := m.get(ctx, rootInodeID)
	if err != nil {
		return nil, err
	}

	for _, segment := range filepath.SplitList(path) {
		if !inode.IsDir {
			return nil, os.ErrInvalid
		}

		i := sort.Search(len(inode.Dirents), func(i int) bool {
			return inode.Dirents[i].Name < segment
		})
		if i >= len(inode.Dirents) || inode.Dirents[i].Name != segment {
			return nil, os.ErrNotExist
		}

		inode, err = m.get(ctx, inode.Dirents[i].Id)
		if err != nil {
			return nil, err
		}
	}

	return inode, nil
}
