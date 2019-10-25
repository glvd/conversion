package conversion

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/multiformats/go-multiaddr"

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

// DefaultNode ...
var DefaultNode = "/ip4/127.0.0.1/tcp/5001"
var _cli *httpapi.HttpApi
var _myID *PeerID

func init() {
	bytes, e := ioutil.ReadFile(os.Getenv("IPFS_PATH"))
	if e != nil {
		return
	}
	DefaultNode = strings.TrimSpace(string(bytes))
}

// MyID ...
func MyID() *PeerID {
	if _myID == nil {
		pid := &PeerID{}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		e := _cli.Request("id").Exec(ctx, pid)
		if e != nil {
			log.Error(e)
			return nil
		}
		_myID = pid
	}
	return _myID
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

// ConnectToNode ...
func ConnectToNode() (e error) {
	ma, err := multiaddr.NewMultiaddr(DefaultNode)
	if err != nil {
		return err
	}
	_cli, e = httpapi.NewApi(ma)
	return
}

// CheckNode ...
func CheckNode() bool {
	return MyID() != nil
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
func AddFile(ctx context.Context, filename string) (string, error) {
	file, e := os.Open(filename)
	if e != nil {
		return "", e
	}
	resolved, e := _cli.Unixfs().Add(ctx, files.NewReaderFile(file),
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
func AddDir(ctx context.Context, dir string) (string, error) {
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
	resolved, e := _cli.Unixfs().Add(ctx, sf,
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
func PinHash(ctx context.Context, hash string) error {
	return _cli.Pin().Add(ctx, path.New(hash), func(settings *options.PinAddSettings) error {
		settings.Recursive = true
		return nil
	})
}

// UnpinHash ...
func UnpinHash(ctx context.Context, hash string) error {
	return _cli.Pin().Rm(ctx, path.New(hash), func(settings *options.PinRmSettings) error {
		settings.Recursive = true
		return nil
	})
}

// PinCheck ...
func PinCheck(ctx context.Context, hash ...string) (int, error) {
	pins, e := _cli.Pin().Ls(ctx, func(settings *options.PinLsSettings) error {
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
