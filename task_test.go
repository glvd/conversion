package conversion

import "testing"

// TestTask_Start ...
func TestTask_Start(t *testing.T) {
	v := NewSourceWalk(&VideoSource{
		Bangumi: "abp-901",
	})
	task := NewTask()
	log.Info("running")
	e := task.AddWalker(v)
	if e != nil {
		t.Fatal(e)
	}
	e = task.Start()
	if e != nil {
		t.Fatal(e)
	}
}
