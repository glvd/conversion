package conversion

import "testing"

func TestGetFileIndex(t *testing.T) {
	if GetFileIndex("abc-123@A") != 1 {
		t.Failed()
	}
}
