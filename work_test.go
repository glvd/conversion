package conversion

import "testing"

// TestGetFileIndex ...
func TestGetFileIndex(t *testing.T) {
	if GetFileIndex("d:\\abc-123@A.mp4") != 1 {
		t.Failed()
	}
	if GetFileIndex("d:\\abc-123@B.mp4") != 2 {
		t.Failed()
	}
	if GetFileIndex("d:\\abc-123.mp4") != 1 {
		t.Failed()
	}
}
