package conversion

import "testing"

func init() {
	SetNodePath(`D:\workspace\ipfs`)
	if err := connectToNode(); err != nil {
		log.Error(err)
	}
}

// TestNodeID ...
func TestNodeID(t *testing.T) {
	if CheckNode() {
		t.Failed()
	}
	t.Logf("%+v", _node.MyID())
}
