package conversion

import "testing"

func init() {

}

// TestNodeID ...
func TestNodeID(t *testing.T) {
	if CheckNode() {
		t.Failed()
	}
	//t.Logf("%+v", _node.MyID())
}
