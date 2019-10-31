package conversion

import (
	"context"
	"errors"
	"strings"

	"github.com/ipfs/interface-go-ipfs-core/path"
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

// CheckNode ...
func CheckNode() bool {
	return globalNode.ID() != nil
}

// CidHash ...
func CidHash(path path.Resolved) string {
	return path.Cid().String()
}

// ResolvedHash ...
func ResolvedHash(path path.Resolved) (string, error) {
	ss := strings.Split(path.String(), "/")
	if len(ss) == 3 {
		return ss[2], nil
	}
	return "", errors.New("wrong resolved data")
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
		log.Infow("node registerd", "type", node.Type())
		globalNode = node
	}
}
