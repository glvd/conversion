package conversion

import "testing"

func init() {
	RegisterDatabase(MustDatabase(InitMySQL()))
}

// TestInsertOrUpdate ...
func TestInsertOrUpdate(t *testing.T) {
	i, e := InsertOrUpdate(&Video{})
	if e != nil {
		t.Fatal(e)
	}
}
