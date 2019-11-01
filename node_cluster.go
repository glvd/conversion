package conversion

import (
	"context"
	"fmt"
	"os"
	"sync"

	api "github.com/glvd/cluster-api"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/multiformats/go-multiaddr"
)

type clusterNode struct {
	client   api.Client
	addParam *api.AddParams
	addr     string
	//recursive         bool
	//quiet             bool
	//quieter           bool
	//noStream          bool
	//layout            string
	//wrapWithDirectory bool
	//hidden            bool
	//chunker           string
	//rawLeaves         bool
	//cidVersion        int
	//hash              string
	//local             bool
	//name              string
	//replicationMin    int
	//replicationMax    int
	//metadata          string
	//allocations       string
	//nocopy            bool
	//shard             bool
	//cfg               *api.Config
}

// NewClusterNode ...
func NewClusterNode(addr string) Node {
	node := &clusterNode{
		addr:     addr,
		addParam: api.DefaultAddParams(),
	}
	if err := node.connect(); err != nil {
		panic(err)
	}
	return node
}

func (c *clusterNode) params() api.AddParams {
	return *c.addParam
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
		p := c.params()
		err := c.client.Add(ctx, []string{filename}, &p, out)
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
		p := c.params()
		p.Recursive = true
		err := c.client.AddMultiFile(ctx, files.NewMultiFileReader(d, false), &p, out)
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
	decoded, e := cid.Decode(hash)
	if e != nil {
		return e
	}
	pin, e := c.client.Pin(ctx, decoded, c.addParam.PinOptions)
	if e != nil {
		return e
	}
	log.Infow("pinned", "name", pin.Name, "hash", pin.Cid.String())
	return nil
}

// UnpinHash ...
func (c *clusterNode) UnpinHash(ctx context.Context, hash string) error {
	decoded, e := cid.Decode(hash)
	if e != nil {
		return e
	}
	pin, e := c.client.Unpin(ctx, decoded)
	if e != nil {
		return e
	}
	log.Infow("unpinned", "name", pin.Name, "hash", pin.Cid.String())
	return nil
}

// PinCheck ...
func (c *clusterNode) PinCheck(ctx context.Context, hash ...string) (int, error) {
	for i, h := range hash {
		decoded, e := cid.Decode(h)
		if e != nil {
			return i, e
		}
		info, err := c.client.Status(ctx, decoded, false)
		if err != nil {
			return i, err
		}
		pinned := false
		for k, m := range info.PeerMap {
			if m.Status.Match(api.TrackerStatusPinned) {
				pinned = true
				break
			}
			log.Infow("pincheck", "peer", k, "hash", info.Cid.String(), "status", m.Status.String())
		}
		if !pinned {
			return i, fmt.Errorf("hash[%s] is not pinned", h)
		}
	}
	return len(hash), nil
}

func (c *clusterNode) connect() (e error) {
	a, e := multiaddr.NewMultiaddr(c.addr)
	if e != nil {
		return e
	}
	c.client, e = api.DefaultCluster(&api.Config{
		APIAddr:           a,
		DisableKeepAlives: true,
	})
	return
}
