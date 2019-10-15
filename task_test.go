package conversion

import (
	"testing"
	"time"

	"github.com/godcong/go-trait"
)

// TestTask_Start ...
func TestTask_Start(t *testing.T) {
	task := NewTask()
	for i := 0; i < 1000; i++ {
		id := trait.GenerateRandomString(64)
		v := NewSourceWalk(&VideoSource{
			Bangumi: id,
		})
		e := task.AddWalker(v)
		if e != nil {
			t.Fatal(e)
		}
		time.Sleep(1 * time.Microsecond)
	}
	log.Info("running")
	e := task.Start()
	if e != nil {
		t.Fatal(e)
	}
}
