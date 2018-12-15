package mjölnir

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	proto "github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/clientv3"
)

func (m *Mjölnir) txn(ctx context.Context, f func(*txn) error) error {
	for {
		txn := txn{
			ctx: ctx,
			kv:  m.kv,
		}
		if err := f(&txn); err != nil {
			return err
		}

		_, err := m.kv.Txn(ctx).If(txn.cmps...).Then(txn.ops...).Commit()
		if err == nil {
			return nil
		}
	}
}

type txn struct {
	ctx  context.Context
	kv   clientv3.KV
	cmps []clientv3.Cmp
	ops  []clientv3.Op
}

func (t *txn) lookup(path string) (*inode, error) {
	inode, err := t.get(rootInodeID)
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

		inode, err = t.get(inode.Dirents[i].Id)
		if err != nil {
			return nil, err
		}
	}

	t.cmps = append(t.cmps, clientv3.Compare(clientv3.Version(inode.ID()), "=", inode.version))
	return inode, nil
}

func (t *txn) get(id uint64) (*inode, error) {
	resp, err := t.kv.Get(t.ctx, strconv.FormatUint(id, 16))
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

func (t *txn) put(i *inode, opts ...clientv3.OpOption) error {
	buf, err := i.Marshal()
	if err != nil {
		return err
	}
	t.ops = append(t.ops, clientv3.OpPut(i.ID(), buf, opts...))
	return nil
}

func (t *txn) remote(id uint64, opts ...clientv3.OpOption) {
	t.ops = append(t.ops, clientv3.OpDelete(strconv.FormatUint(id, 16), opts...))
}
