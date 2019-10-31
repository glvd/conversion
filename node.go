package conversion

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	api "github.com/glvd/cluster-api"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/multiformats/go-multiaddr"
)

// NodeTypeCluster ...
const (
	NodeTypeCluster = "cluster"
	NodeTypeSingle  = "single"
	NodeTypeDummy   = "dummy"
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
	addr   string
	client *httpapi.HttpApi
	id     *PeerID
}

type clusterNode struct {
	client            api.Client
	addParam          *api.AddParams
	recursive         bool
	quiet             bool
	quieter           bool
	noStream          bool
	layout            string
	wrapWithDirectory bool
	hidden            bool
	chunker           string
	rawLeaves         bool
	cidVersion        int
	hash              string
	local             bool
	name              string
	replicationMin    int
	replicationMax    int
	metadata          string
	allocations       string
	nocopy            bool
	shard             bool
	cfg               *api.Config
}

type dummyNode struct {
}

// PeerID ...
type PeerID struct {
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ID              string   `json:"ID"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	PublicKey       string   `json:"PublicKey"`
}

var globalNode Node

func init() {
	globalNode = dummyNode{}
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

// NewClusterNode ...
func NewClusterNode(cfg *api.Config) Node {
	node := &clusterNode{
		cfg:      cfg,
		addParam: api.DefaultAddParams(),
	}
	if err := node.connect(); err != nil {
		panic(err)
	}
	return node
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

// Type ...
func (c *clusterNode) Type() string {
	return NodeTypeCluster

}

// ID ...
func (c *clusterNode) ID() *PeerID {
	id, e := c.client.ID(context.Background())
	if e != nil {
		return nil
	}
	var addrs []string
	for _, addr := range id.Addresses {
		addrs = append(addrs, addr.String())
	}
	return &PeerID{
		Addresses:       addrs,
		AgentVersion:    id.IPFS.AgentVersion,
		ID:              id.ID.String(),
		ProtocolVersion: string(id.RPCProtocolVersion),
		//PublicKey:       "",
	}
}

// AddFile ...
func (c *clusterNode) AddFile(ctx context.Context, filename string) (s string, e error) {
	//param := api.DefaultAddParams()
	//p.ReplicationFactorMin = c.Int("replication-min")
	//p.ReplicationFactorMax = c.Int("replication-max")
	out := make(chan *api.AddedOutput)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := c.client.Add(ctx, []string{filename}, c.addParam, out)
		if err != nil {
			e = err
			return
		}
	}()
	for v := range out {
		log.Info(v.Cid.String())
		s = v.Cid.String()
	}
	wg.Wait()
	return s, e
}

// AddDir ...
func (c *clusterNode) AddDir(ctx context.Context, dir string) (s string, e error) {
	stat, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}

	sf, err := files.NewSerialFile(dir, false, stat)
	if err != nil {
		return "", err
	}
	d := files.NewMapDirectory(map[string]files.Node{"": sf}) // unwrapped on the other side

	out := make(chan *api.AddedOutput)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := c.client.AddMultiFile(ctx, files.NewMultiFileReader(d, false), c.addParam, out)
		if err != nil {
			e = err
			return
		}
	}()
	for v := range out {
		log.Info(v.Cid.String())
		s = v.Cid.String()
	}
	wg.Wait()
	return s, e
}

// PinHash ...
func (c *clusterNode) PinHash(ctx context.Context, hash string) error {
	panic("implement me")
}

// UnpinHash ...
func (c *clusterNode) UnpinHash(ctx context.Context, hash string) error {
	panic("implement me")
}

// PinCheck ...
func (c *clusterNode) PinCheck(ctx context.Context, hash ...string) (int, error) {
	panic("implement me")
}

func (c *clusterNode) connect() (e error) {
	c.client, e = api.DefaultCluster(c.cfg)
	return
}

// Type ...
func (d dummyNode) Type() string {
	return NodeTypeDummy
}

// ID ...
func (d dummyNode) ID() *PeerID {
	return &PeerID{
		Addresses:       nil,
		AgentVersion:    "",
		ID:              "this is dummy",
		ProtocolVersion: "",
		PublicKey:       "",
	}
}

// AddFile ...
func (d dummyNode) AddFile(ctx context.Context, filename string) (string, error) {
	log.Infow("dummy", "func", "AddFile")
	return "this is dummy", nil
}

// AddDir ...
func (d dummyNode) AddDir(ctx context.Context, dir string) (string, error) {
	log.Infow("dummy", "func", "AddDir")
	return "this is dummy", nil
}

// PinHash ...
func (d dummyNode) PinHash(ctx context.Context, hash string) error {
	log.Infow("dummy", "func", "PinHash")
	return nil
}

// UnpinHash ...
func (d dummyNode) UnpinHash(ctx context.Context, hash string) error {
	log.Infow("dummy", "func", "UnpinHash")
	return nil
}

// PinCheck ...
func (d dummyNode) PinCheck(ctx context.Context, hash ...string) (int, error) {
	log.Infow("dummy", "func", "PinCheck")
	return 0, nil
}

// RegisterNode ...
func RegisterNode(node Node) {
	if node != nil && globalNode.Type() != NodeTypeDummy {
		globalNode = node
	}
}
