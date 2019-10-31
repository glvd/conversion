package conversion

import (
	"context"
	"os"
	"time"

	files "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/multiformats/go-multiaddr"
)

// singleNode ...
type singleNode struct {
	addr   string
	client *httpapi.HttpApi
	id     *PeerID
}

// Type ...
func (n *singleNode) Type() string {
	return NodeTypeSingle
}

// ID ...
func (n *singleNode) ID() *PeerID {
	if n.id == nil {
		pid := &PeerID{}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		e := n.client.Request("id").Exec(ctx, pid)
		if e != nil {
			log.Error(e)
			return nil
		}
		n.id = pid
	}
	return n.id
}

// connectToNode ...
func (n *singleNode) connect() (e error) {
	ma, err := multiaddr.NewMultiaddr(n.addr)
	if err != nil {
		return err
	}
	n.client, e = httpapi.NewApi(ma)
	return
}

// NewSingleNode ...
func NewSingleNode(addr string) Node {
	node := &singleNode{addr: addr}
	if err := node.connect(); err != nil {
		panic(err)
	}
	return node
}

// AddFile ...
func (n *singleNode) AddFile(ctx context.Context, filename string) (string, error) {
	file, e := os.Open(filename)
	if e != nil {
		return "", e
	}
	resolved, e := n.client.Unixfs().Add(ctx, files.NewReaderFile(file),
		func(settings *options.UnixfsAddSettings) error {
			settings.Pin = true
			return nil
		})
	if e != nil {
		return "", e
	}
	return CidHash(resolved), nil
}

// AddDir ...
func (n *singleNode) AddDir(ctx context.Context, dir string) (string, error) {
	stat, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}

	sf, err := files.NewSerialFile(dir, false, stat)
	if err != nil {
		return "", err
	}
	//不加目录
	//slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry(filepath.Base(dir), sf)})
	//reader := files.NewMultiFileReader(slf, true)
	resolved, e := n.client.Unixfs().Add(ctx, sf,
		func(settings *options.UnixfsAddSettings) error {
			settings.Pin = true
			return nil
		})
	if e != nil {
		return "", e
	}

	return CidHash(resolved), nil
}

// PinHash ...
func (n *singleNode) PinHash(ctx context.Context, hash string) error {
	return n.client.Pin().Add(ctx, path.New(hash), func(settings *options.PinAddSettings) error {
		settings.Recursive = true
		return nil
	})
}

// UnpinHash ...
func (n *singleNode) UnpinHash(ctx context.Context, hash string) error {
	return n.client.Pin().Rm(ctx, path.New(hash), func(settings *options.PinRmSettings) error {
		settings.Recursive = true
		return nil
	})
}

// PinCheck ...
func (n *singleNode) PinCheck(ctx context.Context, hash ...string) (int, error) {
	pins, e := n.client.Pin().Ls(ctx, func(settings *options.PinLsSettings) error {
		settings.Type = "recursive"
		return nil
	})
	if e != nil {
		return -1, e
	}
	var ps []string
	for _, pin := range pins {
		ps = append(ps, CidHash(pin.Path()))
	}
	log.Infow("pincheck", "hash", ps)
	for i, v := range hash {
		if !ExistVerifyString(v, ps...) {
			return i, nil
		}
	}
	return len(hash), nil
}
