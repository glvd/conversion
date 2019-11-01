package conversion

import (
	"context"
	"testing"
	"time"
)

func init() {

}

// TestNewSingleNode ...
func TestNewSingleNode(t *testing.T) {
	node := NewSingleNode("/ip4/127.0.0.1/tcp/5001")
	s, err := node.AddDir(context.Background(), "D:\\temp\\ca7946ec-eeb6-4a03-a2d2-e47ab343a934")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
	err = node.PinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}

	i, err := node.PinCheck(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(i)

	err = node.UnpinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	i, err = node.PinCheck(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(i)
}

// TestNodeID ...
func TestNewClusterNode(t *testing.T) {
	node := NewClusterNode("/ip4/127.0.0.1/tcp/9094")

	s, err := node.AddDir(context.Background(), "D:\\temp\\ca7946ec-eeb6-4a03-a2d2-e47ab343a934")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)

	err = node.PinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	//s := "QmbVToPV7VmozhFZaLujMxF6VNP5v32pN4iRg1BX76QmSE"
	i, err := node.PinCheck(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(i)

	err = node.UnpinHash(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	i, err = node.PinCheck(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(i)
}
