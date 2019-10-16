package conversion

import (
	"testing"
	"time"

	"github.com/gotrait/tool"
)

// TestTask_Start ...
func TestTask_Start(t *testing.T) {
	task := NewTask()
	for i := 0; i < 100; i++ {
		id := tool.GenerateRandomString(2)
		v, e := NewSourceWalk(&VideoSource{
			Bangumi: id,
		})
		if e != nil {
			t.Fatal(e)
		}
		e = task.AddWalker(v)
		if e != nil {
			t.Fatal(e)
		}
		id = tool.GenerateRandomString(2)
		v1, e1 := NewInfoWalk(&VideoInfo{
			ID: id,
		})
		if e1 != nil {
			t.Fatal(e1)
		}
		e = task.AddWalker(v1)
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
