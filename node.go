package conversion

import (
	"context"
	"errors"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"os"
	"path/filepath"
	"strings"
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
var _cli *httpapi.HttpApi
var _myID *PeerID

func init() {
	_node = os.Getenv("IPFS_PATH")
}

// MyID ...
func MyID() *PeerID {
	if _myID == nil {
		_myID = &PeerID{}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		e := _cli.Request("id").Exec(ctx, _myID)
		if e != nil {
			log.Error(e)
			return nil
		}
	}
	return _myID
}

// SetNodePath ...
func SetNodePath(path string) {
	_node = path
}

// ConnectToNode ...
func ConnectToNode() (e error) {
	_cli, e = httpapi.NewPathApi(_node)
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
	for i, v := range hash {
		for _, pin := range pins {
			if hash, err := ResolvedHash(pin.Path()); err != nil {
				return i, err
				if hash == v {
					break
				}
			}
			return i, nil
		}
	}
	return len(hash), nil
}
