package ctl

import (
	"testing"
)

func TestFileWrite(t *testing.T) {
	f := New("ctl")
	go f.Write([]byte("hello"))
	t.Log(<-f.Out())
}

func TestFileRead(t *testing.T) {
	f := New("ctl")
	f.In() <- "hello"
	b := make([]byte, 5)
	f.Read(b)
	t.Log(string(b))
}