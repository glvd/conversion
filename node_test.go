package conversion

import (
	"context"
	"testing"

	api "github.com/glvd/cluster-api"
	"github.com/multiformats/go-multiaddr"
)

func init() {

}

// TestNewSingleNode ...
func TestNewSingleNode(t *testing.T) {
	node := NewSingleNode("/ip4/127.0.0.1/tcp/5001")
	s, err := node.AddDir(context.Background(), "D:\\temp\\ca7946ec-eeb6-4a03-a2d2-e47ab343a934")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
	err = node.PinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}

	i, err := node.PinCheck(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(i)

	err = node.UnpinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
}

// TestNodeID ...
func TestNewClusterNode(t *testing.T) {
	a, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9094")
	if err != nil {
		t.Fatal(err)
	}
	node := NewClusterNode(&api.Config{
		APIAddr:           a,
		DisableKeepAlives: true,
	})

	s, err := node.AddDir(context.Background(), "D:\\temp\\ca7946ec-eeb6-4a03-a2d2-e47ab343a934")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)

	err = node.PinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}

	err = node.UnpinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
}
