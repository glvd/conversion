package conversion

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/ipfs/go-ipfs-http-client"
)

// PeerID ...
type PeerID struct {
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ID              string   `json:"ID"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	PublicKey       string   `json:"PublicKey"`
}

var _node string
var cli *httpapi.HttpApi

func init() {
	_node = os.Getenv("IPFS_PATH")
}

// NodeID ...
func NodeID() *PeerID {
	var nid PeerID
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	e := cli.Request("id").Exec(ctx, &nid)
	if e != nil {
		log.Error(e)
		return nil
	}
	return &nid
}

// SetNodePath ...
func SetNodePath(path string) {
	_node = path
}

// ConnectToNode ...
func ConnectToNode() (e error) {
	cli, e = httpapi.NewPathApi(_node)
	return
}

// CheckNode ...
func CheckNode() bool {
	info, e := os.Stat(filepath.Join(_node, "api"))
	if e != nil || info.IsDir() {
		return false
	}
	return true
}
