package conversion

import (
	"testing"
	"time"

	_ "github.com/gocacher/badger-cache/easy"
	"github.com/gotrait/tool"
)

// TestTask_Start ...
func TestTask_Start(t *testing.T) {

	task := NewTask()
	for i := 0; i < 5; i++ {
		id := tool.GenerateRandomString(5)
		v, e := NewSourceWork(&VideoPornSource{
			VideoPath: []string{"D:\\video\\demo-r-24.mp4"},
			Bangumi:   id,
		})
		if e != nil {
			t.Fatal(e)
		}
		e = task.AddWorker(v, false)
		if e != nil {
			t.Fatal(e)
		}
		id = tool.GenerateRandomString(2)
		v1, e1 := NewInfoWork(&VideoPornInfo{
			ID: id,
		})
		if e1 != nil {
			t.Fatal(e1)
		}
		e = task.AddWorker(v1, false)
		if e != nil {
			t.Fatal(e)
		}
		time.Sleep(1 * time.Microsecond)
	}
	log.Info("running")
	task.Limit = 2
	go func() {
		time.Sleep(30 * time.Second)
		//task.Stop()
	}()
	e := task.Start()
	if e != nil {
		t.Fatal(e)
	}
}
