package conversion

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/glvd/cluster/api/rest/client"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/multiformats/go-multiaddr"

	"github.com/ipfs/go-ipfs-http-client"
)

// NodeTypeCluster ...
const (
	NodeTypeCluster = "cluster"
	NodeTypeSingle  = "single"
)

// Node ...
type Node interface {
	Type() string
	ID() *PeerID
	AddFile(ctx context.Context, filename string) (string, error)
	AddDir(ctx context.Context, dir string) (string, error)
	PinHash(ctx context.Context, hash string) error
	UnpinHash(ctx context.Context, hash string) error
	PinCheck(ctx context.Context, hash ...string) (int, error)
}

// singleNode ...
type singleNode struct {
	client *httpapi.HttpApi
	id     *PeerID
}

type clusterNode struct {
	client client.Client
}

// PeerID ...
type PeerID struct {
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ID              string   `json:"ID"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	PublicKey       string   `json:"PublicKey"`
}

// DefaultNode ...
var DefaultNode = "/ip4/127.0.0.1/tcp/5001"
var defaultNode Node

func init() {
	bytes, e := ioutil.ReadFile(os.Getenv("IPFS_PATH"))
	if e != nil {
		return
	}
	DefaultNode = strings.TrimSpace(string(bytes))
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

// SetNodePath ...
func SetNodePath(path string) {
	bytes, e := ioutil.ReadFile(path)
	if e != nil {
		return
	}
	DefaultNode = strings.TrimSpace(string(bytes))
}

// SetNodeAddress ...
func SetNodeAddress(addr string) {
	DefaultNode = addr
}

// connectToNode ...
func (n *singleNode) connect() (e error) {
	ma, err := multiaddr.NewMultiaddr(DefaultNode)
	if err != nil {
		return err
	}
	n.client, e = httpapi.NewApi(ma)
	return
}

// CheckNode ...
func CheckNode() bool {
	panic("todo")
}

// ResolvedHash ...
func ResolvedHash(path path.Resolved) (string, error) {
	ss := strings.Split(path.String(), "/")
	if len(ss) == 3 {
		return ss[2], nil
	}
	return "", errors.New("wrong resolved data")
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
	return ResolvedHash(resolved)
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

	return ResolvedHash(resolved)
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
	var h string
	for _, pin := range pins {
		if h, e = ResolvedHash(pin.Path()); e != nil {
			return 0, e
		}
		ps = append(ps, h)
	}

	for i, v := range hash {
		if !ExistVerifyString(v, ps...) {
			return i, nil
		}
	}
	return len(hash), nil
}

// InitNode ...
func initNode() error {
	_node = &singleNode{
		ModeType: "single",
	}
	if err := connectToNode(); err != nil {
		return err
	}
	_node.ID = _node.MyID()
	return nil
}
