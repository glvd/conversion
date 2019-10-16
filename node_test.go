package conversion

import "testing"

func init() {
	SetNodePath(`D:\workspace\ipfs`)
	if err := ConnectToNode(); err != nil {
		log.Error(err)
	}
}

// TestNodeID ...
func TestNodeID(t *testing.T) {
	if CheckNode() {
		t.Failed()
	}
	t.Logf("%+v", MyID())
}
