package conversion

import "testing"

func init() {
	RegisterDatabase(MustDatabase(InitMySQL()))
	e := SyncTable()
	if e != nil {
		panic(e)
	}
}

// TestInsertOrUpdate ...
func TestInsertOrUpdate(t *testing.T) {
	i, e := InsertOrUpdate(&Video{})
	if e != nil {
		t.Fatal(e)
	}
	if i == 0 {
		t.Failed()
	}
}
