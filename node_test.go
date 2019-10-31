package conversion

import (
	"context"
	"testing"

	api "github.com/glvd/cluster-api"
	"github.com/multiformats/go-multiaddr"
)

func init() {

}

// TestNodeID ...
func TestNodeID(t *testing.T) {
	a, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9094")
	if err != nil {
		t.Fatal(err)
	}
	node := NewClusterNode(&api.Config{
		SSL:               false,
		NoVerifyCert:      false,
		Username:          "",
		Password:          "",
		APIAddr:           a,
		Host:              "",
		Port:              "9094",
		ProtectorKey:      nil,
		ProxyAddr:         nil,
		Timeout:           0,
		DisableKeepAlives: true,
		LogLevel:          "",
	})

	s, err := node.AddDir(context.Background(), "D:\\temp\\ca7946ec-eeb6-4a03-a2d2-e47ab343a934")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
}
