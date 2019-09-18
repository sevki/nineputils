package ctl

import (
	"bufio"
	"bytes"
	"io"
	"net/textproto"
	"os"
	"time"
)

func New(name string) File {
	c := &ctlfile{
		name: name,
		sync: make(chan interface{}),
		in:   make(chan string),
		out:  make(chan string),
	}
	r := textproto.NewReader(bufio.NewReader(&c.inbuf))
	w := bufio.NewWriter(&c.outbuf)
	go readloop(r, c.sync, c.out)
	go writeloop(w, c.in)
	return c
}

type ctlfile struct {
	name          string
	inbuf, outbuf bytes.Buffer
	sync          chan interface{}
	in, out       chan string
}

type File interface {
	os.FileInfo
	io.ReadWriteCloser
	In() chan<- string
	Out() <-chan string
}

func (c *ctlfile) Name() string       { return c.name }
func (c *ctlfile) Size() int64        { return 0 }
func (c *ctlfile) Mode() os.FileMode  { return 0777 }
func (c *ctlfile) ModTime() time.Time { return time.Now().Truncate(time.Hour) }
func (c *ctlfile) IsDir() bool        { return false }
func (c *ctlfile) Sys() interface{}   { return c }

// Read implements the read interface on ctlfile
// which means it reads from the outbuffer
func (c *ctlfile) Read(p []byte) (n int, err error) {
	return c.outbuf.Read(p)
}

// Write implmenets the io.Write interface on ctlfile
// which means it writes to the inbuffer, and sends a nil
// message to insync channel
func (c *ctlfile) Write(p []byte) (n int, err error) {
	defer func() { c.sync <- nil }()
	return c.inbuf.Write(p)
}

func (c *ctlfile) Close() error {
	close(c.sync)
	return nil
}

func (c *ctlfile) In() chan<- string  { return c.in }
func (c *ctlfile) Out() <-chan string { return c.out }
func readloop(r *textproto.Reader, sync chan interface{}, out chan string) {
	for {
		_, ok := <-sync
		if !ok {
			return
		}
		s, err := r.ReadLine()
		if err == io.EOF {
			continue
		} else if err != nil {
			return
		}
		out <- s
	}
}

func writeloop(w *bufio.Writer, in chan string) {
	for {
		s := <-in
		w.Write([]byte(s))
		w.Flush()
	}

}