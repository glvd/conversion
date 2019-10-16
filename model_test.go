package conversion

import (
	"testing"

	"github.com/go-xorm/xorm"
)

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

// TestFindAll ...
func TestFindAll(t *testing.T) {
	e := FindAll(&Video{}, func(rows *xorm.Rows) error {
		var v Video
		if err := rows.Scan(&v); err != nil {
			return err
		}
		log.Info(v)
		return nil
	}, 2, 0)
	if e != nil {
		t.Fatal(e)
	}
}
